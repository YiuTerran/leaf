package udp

import (
	"net"
	"sync"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/processor"
	"github.com/YiuTerran/leaf/util/leafutil"
	"github.com/YiuTerran/leaf/util/netutil"
)

type ReceivedContext struct {
	Addr   net.Addr
	Server *Server
}

type MsgInfo struct {
	Addr net.Addr
	Msg  []byte
}

type Server struct {
	Addr       string
	BufferSize int
	Processor  processor.Processor
	MaxTry     int

	closeSig  chan struct{}
	readChan  chan *MsgInfo
	writeChan chan *MsgInfo
	conn      net.PacketConn
	wg        *sync.WaitGroup
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
	server.closeSig = make(chan struct{}, 1)
	server.writeChan = make(chan *MsgInfo, server.BufferSize)
	server.readChan = make(chan *MsgInfo, server.BufferSize)
	server.conn = conn
	server.wg = &sync.WaitGroup{}
	go server.listen()
	go server.doWrite()
	go server.doRead()
	server.wg.Add(3)
}

func (server *Server) WriteMsg(msg interface{}, addr net.Addr) error {
	if len(server.writeChan) == cap(server.writeChan) {
		return ChanFullError
	}
	if bs, err := server.Processor.Marshal(msg); err != nil {
		return err
	} else {
		server.writeChan <- &MsgInfo{
			Addr: addr,
			Msg:  netutil.MergeBytes(bs),
		}
	}
	return nil
}

func (server *Server) doWrite() {
	defer leafutil.RecoverFromPanic(nil)
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
	server.wg.Done()
}

func (server *Server) doRead() {
	defer leafutil.RecoverFromPanic(nil)
	for b := range server.readChan {
		if b == nil {
			break
		}
		msg, err := server.Processor.Unmarshal(b.Msg)
		if err != nil {
			log.Error("fail to decode udp msg:%v", err)
			continue
		}
		err = server.Processor.Route(msg, &ReceivedContext{
			Addr:   b.Addr,
			Server: server,
		})
		if err != nil {
			log.Error("fail to route udp msg:%v", err)
			continue
		}
	}
	server.wg.Done()
}

func (server *Server) listen() {
	defer leafutil.RecoverFromPanic(nil)
	for {
		select {
		case <-server.closeSig:
			server.writeChan <- nil
			server.readChan <- nil
			server.wg.Done()
			return
		default:
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
}

func (server *Server) Close() {
	_ = server.conn.Close()
	server.closeSig <- struct{}{}
}

func (server *Server) CloseAndWait() {
	server.Close()
	server.wg.Wait()
}
