package ws

import (
	"sync"
	"time"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/network"
	"github.com/gorilla/websocket"
)

type Client struct {
	sync.Mutex
	Addr             string
	ConnNum          int
	ConnectInterval  time.Duration
	PendingWriteNum  int
	MaxMsgLen        uint32
	HandshakeTimeout time.Duration
	AutoReconnect    bool
	NewAgent         func(*Conn) network.Agent
	TextFormat       bool

	dialer    websocket.Dialer
	conns     WebsocketConnSet
	wg        sync.WaitGroup
	closeFlag bool
}

func (client *Client) Start() {
	client.init()

	for i := 0; i < client.ConnNum; i++ {
		client.wg.Add(1)
		go client.connect()
	}
}

func (client *Client) init() {
	client.Lock()
	defer client.Unlock()

	if client.ConnNum <= 0 {
		client.ConnNum = 1
		log.Debug("invalid ConnNum, reset to %v", client.ConnNum)
	}
	if client.ConnectInterval <= 0 {
		client.ConnectInterval = 3 * time.Second
		log.Debug("invalid ConnectInterval, reset to %v", client.ConnectInterval)
	}
	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 100
		log.Debug("invalid BufferSize, reset to %v", client.PendingWriteNum)
	}
	if client.MaxMsgLen <= 0 {
		client.MaxMsgLen = 2048000
		log.Debug("invalid MaxMsgLen, reset to %v", client.MaxMsgLen)
	}
	if client.HandshakeTimeout <= 0 {
		client.HandshakeTimeout = 10 * time.Second
		log.Debug("invalid HandshakeTimeout, reset to %v", client.HandshakeTimeout)
	}
	if client.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
	if client.conns != nil {
		log.Fatal("client is running")
	}

	client.conns = make(WebsocketConnSet)
	client.closeFlag = false
	client.dialer = websocket.Dialer{
		HandshakeTimeout: client.HandshakeTimeout,
	}
}

func (client *Client) dial() *websocket.Conn {
	for {
		conn, _, err := client.dialer.Dial(client.Addr, nil)
		if err == nil || client.closeFlag {
			return conn
		}

		log.Info("connect to %v error: %v", client.Addr, err)
		time.Sleep(client.ConnectInterval)
		continue
	}
}

func (client *Client) connect() {
	defer client.wg.Done()

reconnect:
	conn := client.dial()
	if conn == nil {
		return
	}
	conn.SetReadLimit(int64(client.MaxMsgLen))

	client.Lock()
	if client.closeFlag {
		client.Unlock()
		_ = conn.Close()
		return
	}
	client.conns[conn] = struct{}{}
	client.Unlock()

	wsConn := newWSConn(conn, client.PendingWriteNum, client.MaxMsgLen, client.TextFormat)
	agent := client.NewAgent(wsConn)
	agent.Run()

	// cleanup
	wsConn.Close()
	client.Lock()
	delete(client.conns, conn)
	client.Unlock()
	agent.OnClose()

	if client.AutoReconnect {
		time.Sleep(client.ConnectInterval)
		goto reconnect
	}
}

func (client *Client) Close() {
	client.Lock()
	client.closeFlag = true
	for conn := range client.conns {
		_ = conn.Close()
	}
	client.conns = nil
	client.Unlock()

	client.wg.Wait()
}
