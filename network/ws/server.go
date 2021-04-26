package ws

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/network"
	"github.com/gorilla/websocket"
)

type Server struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	HTTPTimeout     time.Duration
	CertFile        string
	KeyFile         string
	NewAgent        func(*Conn) network.Agent
	AuthFunc        func(*http.Request) (bool, interface{})
	TextFormat      bool //纯文本还是二进制

	ln      net.Listener
	handler *Handler
}

type Handler struct {
	textFormat      bool
	maxConnNum      int
	pendingWriteNum int
	authFunc        func(*http.Request) (bool, interface{})
	maxMsgLen       uint32
	newAgent        func(*Conn) network.Agent
	upgrader        websocket.Upgrader
	conns           WebsocketConnSet
	mutexConns      sync.Mutex
	wg              sync.WaitGroup
}

type Option func(*Server)

func NewServer(port int, newAgentCallback func(*Conn) network.Agent, options ...Option) *Server {
	if port <= 0 || port > 65535 {
		return nil
	}
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	server := &Server{
		Addr:            addr,
		MaxConnNum:      20000,
		PendingWriteNum: 1000,
		MaxMsgLen:       1024000,
		HTTPTimeout:     10 * time.Second,
		NewAgent:        newAgentCallback,
		TextFormat:      false,
	}
	for _, option := range options {
		option(server)
	}
	return server
}

func WithMaxConnNum(num int) Option {
	return func(server *Server) {
		server.MaxConnNum = num
	}
}

func WithPendingWriteNum(num int) Option {
	return func(server *Server) {
		server.PendingWriteNum = num
	}
}

func WithMaxMsgLen(num uint32) Option {
	return func(server *Server) {
		server.MaxMsgLen = num
	}
}

func WithHttpTimeout(duration time.Duration) Option {
	return func(server *Server) {
		server.HTTPTimeout = duration
	}
}

func WithHttpsCert(cert, key string) Option {
	return func(server *Server) {
		server.CertFile = cert
		server.KeyFile = key
	}
}

func WithAuthFunc(authFunc func(*http.Request) (bool, interface{})) Option {
	return func(server *Server) {
		server.AuthFunc = authFunc
	}
}

func WithTextFormat(usingText bool) Option {
	return func(server *Server) {
		server.TextFormat = usingText
	}
}

func getRealIP(req *http.Request) net.Addr {
	ip := req.Header.Get("X-FORWARDED-FOR")
	if ip == "" {
		ip = req.Header.Get("X-REAL-IP")
	}
	if ip != "" {
		ip = strings.Split(ip, ",")[0]
	} else {
		ip, _, _ = net.SplitHostPort(req.RemoteAddr)
	}
	q := net.ParseIP(ip)
	addr := &net.IPAddr{IP: q}
	return addr
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var (
		ok       bool
		userData interface{}
	)
	if handler.authFunc != nil {
		if ok, userData = handler.authFunc(r); !ok {
			http.Error(w, "Forbidden", 403)
		}
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.maxMsgLen))

	handler.wg.Add(1)
	defer handler.wg.Done()

	handler.mutexConns.Lock()
	if handler.conns == nil {
		handler.mutexConns.Unlock()
		_ = conn.Close()
		return
	}
	if len(handler.conns) >= handler.maxConnNum {
		handler.mutexConns.Unlock()
		_ = conn.Close()
		log.Debug("too many connections")
		return
	}
	handler.conns[conn] = struct{}{}
	handler.mutexConns.Unlock()

	wsConn := newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen, handler.textFormat)
	wsConn.remoteOriginIP = getRealIP(r)
	wsConn.userData = userData
	agent := handler.newAgent(wsConn)
	agent.Run()

	// cleanup
	wsConn.Close()
	handler.mutexConns.Lock()
	delete(handler.conns, conn)
	handler.mutexConns.Unlock()
	agent.OnClose()
}

func (server *Server) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("fail to start tcp server: %v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Info("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Info("invalid BufferSize, reset to %v", server.PendingWriteNum)
	}
	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 1024000
		log.Info("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Info("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	if server.CertFile != "" || server.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err != nil {
			log.Fatal("%v", err)
		}

		ln = tls.NewListener(ln, config)
	}

	server.ln = ln
	server.handler = &Handler{
		textFormat:      server.TextFormat,
		authFunc:        server.AuthFunc,
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		maxMsgLen:       server.MaxMsgLen,
		newAgent:        server.NewAgent,
		conns:           make(WebsocketConnSet),
		upgrader: websocket.Upgrader{
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        server.handler,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go func() {
		if err = httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatal("fail to start websocket server:%v", err)
		}
	}()
}

func (server *Server) Close() {
	_ = server.ln.Close()

	server.handler.mutexConns.Lock()
	for conn := range server.handler.conns {
		_ = conn.Close()
	}
	server.handler.conns = nil
	server.handler.mutexConns.Unlock()

	server.handler.wg.Wait()
}
