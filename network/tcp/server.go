package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/network"
)

type Server struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	NewAgent        func(*Conn) network.Agent
	ln              net.Listener
	conns           ConnSet
	mutexConns      sync.Mutex
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup

	// msg parser
	Parser IParser
}

func (server *Server) Start() {
	server.init()
	go server.run()
}

func (server *Server) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal("fail to start tcp server:%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	server.ln = ln
	server.conns = make(ConnSet)

	// msg parser
	if server.Parser == nil {
		msgParser := NewDefaultParser()
		server.Parser = msgParser
	}
}

func (server *Server) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Info("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			_ = conn.Close()
			log.Debug("too many connections")
			continue
		}
		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()

		server.wgConns.Add(1)

		tcpConn := newConn(conn, server.PendingWriteNum, server.Parser)
		agent := server.NewAgent(tcpConn)
		go func() {
			agent.Run()

			// cleanup
			tcpConn.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()
			agent.OnClose()

			server.wgConns.Done()
		}()
	}
}

func (server *Server) Close() {
	_ = server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.conns {
		_ = conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()
	server.wgConns.Wait()
}
