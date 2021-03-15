package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/network"
)

type Client struct {
	sync.Mutex
	Addr            string
	ConnNum         int
	ConnectInterval time.Duration
	PendingWriteNum int
	AutoReconnect   bool
	NewAgent        func(*Conn) network.Agent
	Parser          IParser

	conns     ConnSet
	wg        sync.WaitGroup
	closeFlag bool
}

type Option func(*Client)

func NewClient(addr string, newAgentCallback func(*Conn) network.Agent, options ...Option) *Client {
	c := &Client{
		Addr:            addr,
		ConnNum:         1,
		ConnectInterval: 3 * time.Second,
		AutoReconnect:   true,
		PendingWriteNum: 100,
		Parser:          NewDefaultParser(),
		NewAgent:        newAgentCallback,
		conns:           ConnSet{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

func ConnNum(num int) Option {
	return func(client *Client) {
		client.ConnNum = num
	}
}

func ConnectInterval(dr time.Duration) Option {
	return func(client *Client) {
		client.ConnectInterval = dr
	}
}

func Parser(p IParser) Option {
	return func(client *Client) {
		client.Parser = p
	}
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
	if client.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
	if client.conns != nil {
		log.Fatal("client is running")
	}

	client.conns = make(ConnSet)
	client.closeFlag = false

	if client.Parser == nil {
		// msg parser
		msgParser := NewDefaultParser()
		client.Parser = msgParser
	}
}

func (client *Client) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", client.Addr)
		if err == nil || client.closeFlag {
			return conn
		}

		log.Error("connect to %v error: %v", client.Addr, err)
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

	client.Lock()
	if client.closeFlag {
		client.Unlock()
		_ = conn.Close()
		return
	}
	client.conns[conn] = struct{}{}
	client.Unlock()

	tcpConn := newConn(conn, client.PendingWriteNum, client.Parser)
	agent := client.NewAgent(tcpConn)
	agent.Run()

	// cleanup
	tcpConn.Close()
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
