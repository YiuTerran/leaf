package udp

import (
	"net"
	"sync"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/processor"
	"github.com/YiuTerran/leaf/util/leafutil"
	"github.com/YiuTerran/leaf/util/netutil"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"
)

//这是一个异步的udp客户端，代码和服务端很类似，这里调用了DialUDP，因此视为"有连接的"
//如果需要做一些同步操作，最好使用Client，而不是这个版本

const (
	NotInit = 0
	Inited  = 1
	Closed  = 2
)

type AsyncClient struct {
	ServerAddr string
	BufferSize int
	MaxTry     int //最多尝试次数
	Processor  processor.Processor

	closeSig  chan struct{}
	writeChan chan []byte
	readChan  chan []byte
	conn      *net.UDPConn
	wg        *sync.WaitGroup
	status    atomic.Int32
}

func (client *AsyncClient) Start() error {
	if !client.status.CAS(NotInit, Inited) {
		return xerrors.New("server inited")
	}
	if client.MaxTry <= 0 {
		client.MaxTry = 3
	}
	if client.BufferSize <= 0 {
		client.BufferSize = 100
	}
	if client.Processor == nil {
		log.Fatal("udp client no processor registered!")
	}
	client.closeSig = make(chan struct{}, 1)
	client.writeChan = make(chan []byte, client.BufferSize)
	client.readChan = make(chan []byte, client.BufferSize)
	client.wg = &sync.WaitGroup{}

	client.conn = client.dial()
	if client.conn == nil {
		return InitError
	}
	go client.doWrite()
	go client.listen()
	go client.doRead()
	client.wg.Add(3)
	return nil
}

func (client *AsyncClient) listen() {
	defer leafutil.RecoverFromPanic(nil)
	for {
		select {
		case <-client.closeSig:
			client.readChan <- nil
			client.writeChan <- nil
			client.wg.Done()
			return
		default:
			buffer := make([]byte, DefaultPacketSize)
			n, err := client.conn.Read(buffer)
			if err != nil {
				continue
			}
			buffer = buffer[:n]
			client.readChan <- buffer
		}
	}
}

func (client *AsyncClient) doWrite() {
	defer leafutil.RecoverFromPanic(nil)
	for b := range client.writeChan {
		if b == nil {
			break
		}
		count := client.MaxTry
		for count > 0 {
			_, err := client.conn.Write(b)
			if err != nil {
				log.Error("fail to write udp chan:%+v", err)
			} else {
				break
			}
			count--
		}
	}
	client.wg.Done()
}

func (client *AsyncClient) doRead() {
	defer leafutil.RecoverFromPanic(nil)
	for b := range client.readChan {
		if b == nil {
			break
		}
		msg, err := client.Processor.Unmarshal(b)
		if err != nil {
			log.Error("unable to unmarshal udp msg, ignore")
			continue
		}
		err = client.Processor.Route(msg, client)
		if err != nil {
			log.Error("fail to route udp msg:%v", err)
			continue
		}
	}
	client.wg.Done()
}

func (client *AsyncClient) dial() *net.UDPConn {
	rAddr, err := net.ResolveUDPAddr("udp", client.ServerAddr)
	if err != nil {
		log.Error("fail to resolve add for udp:%v", client.ServerAddr)
		return nil
	}
	conn, err := net.DialUDP("udp", nil, rAddr)
	if err == nil {
		return conn
	}

	log.Error("connect to %v error: %v", client.ServerAddr, err)
	return nil
}

func (client *AsyncClient) WriteMsg(msg interface{}) error {
	args, err := client.Processor.Marshal(msg)
	if err != nil {
		return err
	}
	if client.status.Load() == Closed {
		return xerrors.New("client closed")
	}
	if len(client.writeChan) == cap(client.writeChan) {
		return ChanFullError
	}
	client.writeChan <- netutil.MergeBytes(args)
	return nil
}

func (client *AsyncClient) Close() {
	if !client.status.CAS(Inited, Closed) {
		return
	}
	_ = client.conn.Close()
	client.closeSig <- struct{}{}
}

func (client *AsyncClient) CloseAndWait() {
	client.Close()
	client.wg.Wait()
}

func (client *AsyncClient) Destroy() {
	client.status.Store(Closed)
	_ = client.conn.Close()
	close(client.writeChan)
}
