package udp

import (
	"net"
	"sync"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/processor"
)

type ClientInfo struct {
	Addr   net.Addr
	Server *Server
}

type MsgInfo struct {
	Addr net.Addr
	Msg  []byte
}

type Server struct {
	sync.Mutex
	Addr       string
	BufferSize int
	Processor  processor.Processor
	MaxTry     int

	closeFlag bool
	readChan  chan *MsgInfo
	writeChan chan *MsgInfo
	conn      net.PacketConn
}

func (server *Server) Start() {
	conn, err := net.ListenPacket("udp", server.Addr)
	if err != nil {
		log.Fatal("fail to bind udp port:%v", err)
	}
	if server.MaxTry <= 0 {
		server.MaxTry = 3
	}
	if server.BufferSize <= 0 {
		server.BufferSize = 100
	}
	server.writeChan = make(chan *MsgInfo, server.BufferSize)
	server.readChan = make(chan *MsgInfo, server.BufferSize)
	server.conn = conn
	go server.listen()
	go server.doWrite()
	go server.doRead()
}

func (server *Server) doWrite() {
	for b := range server.writeChan {
		if b == nil {
			break
		}
		count := server.MaxTry
		for count > 0 {
			_, err := server.conn.WriteTo(b.Msg, b.Addr)
			if err != nil {
				log.Error("fail to write udp chan:%+v", err)
				count--
			} else {
				break
			}
		}
	}

	_ = server.conn.Close()
	server.Lock()
	server.closeFlag = true
	server.Unlock()
}

func (server *Server) doRead() {
	for b := range server.readChan {
		if b == nil {
			return
		}
		msg, err := server.Processor.Unmarshal(b.Msg)
		if err != nil {
			log.Error("fail to decode udp msg:%v", err)
			continue
		}
		err = server.Processor.Route(msg, &ClientInfo{
			Addr:   b.Addr,
			Server: server,
		})
		if err != nil {
			log.Error("fail to route udp msg:%v", err)
			continue
		}
	}
}

func (server *Server) listen() {
	for {
		buffer := make([]byte, DefaultPacketSize)
		n, addr, err := server.conn.ReadFrom(buffer)
		if err != nil {
			log.Error("fail to doRead udp msg:%v", err)
			continue
		}
		if len(server.readChan) == cap(server.readChan) {
			log.Error("doRead chan full, drop msg from %v", addr)
			continue
		}
		server.readChan <- &MsgInfo{
			Addr: addr,
			Msg:  buffer[:n],
		}
	}
}

func (server *Server) Close() {
	server.Lock()
	defer server.Unlock()
	if server.closeFlag {
		return
	}
	server.writeChan <- nil
	server.readChan <- nil
	server.closeFlag = true
}

func (server *Server) Destroy() {
	server.Lock()
	defer server.Unlock()
	close(server.readChan)
	close(server.writeChan)
	server.closeFlag = true
}
