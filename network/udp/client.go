package udp

import (
	"net"
	"sync"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/processor"
	"github.com/YiuTerran/leaf/util/netutil"
)

//udp是无连接的，直接将字节流读入读出即可
//操作系统将客户端一次发过来的包一次性放入缓冲区，因此没有所谓的粘包问题；go使用了linux的UDP设计实现，bind本质上只是绑定了一个地址
//但是单次包最大长度是有限制的(理论最大值：65507)，如果超出这个限制，就要拆包；但是由于UDP不可靠，这种情况下就要自己实现TCP的很多功能
//另外，由于MTU的限制，UDP的长度最好在512字节之内（参考http://dwz.win/vN5)
//所以UDP这里的设计直接把processor嵌入进去了，每次发送和接收都是独立的，不过设置了发送缓冲区

type Client struct {
	sync.RWMutex
	ServerAddr string
	BufferSize int
	MaxTry     int //最多尝试次数
	Processor  processor.Processor
	CloseSig   chan struct{}

	writeChan chan []byte
	readChan  chan []byte
	conn      *net.UDPConn
	wg        *sync.WaitGroup
	closeFlag bool
}

func (client *Client) Start() error {
	client.Lock()
	defer client.Unlock()
	client.closeFlag = false

	if client.MaxTry <= 0 {
		client.MaxTry = 3
	}
	if client.BufferSize <= 0 {
		client.BufferSize = 100
	}
	if client.Processor == nil {
		log.Fatal("udp client no processor registered!")
	}
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

func (client *Client) listen() {
	for {
		select {
		case <-client.CloseSig:
			client.readChan <- nil
			client.writeChan <- nil
			client.Lock()
			client.closeFlag = true
			client.Unlock()
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

func (client *Client) doWrite() {
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

func (client *Client) doRead() {
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

func (client *Client) dial() *net.UDPConn {
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

func (client *Client) WriteMsg(msg interface{}) error {
	args, err := client.Processor.Marshal(msg)
	if err != nil {
		return err
	}
	client.RLock()
	defer client.RUnlock()
	if client.closeFlag {
		return nil
	}
	if len(client.writeChan) == cap(client.writeChan) {
		return ChanFullError
	}
	client.writeChan <- netutil.MergeBytes(args)
	return nil
}

func (client *Client) Close() {
	client.Lock()
	defer client.Unlock()
	if client.closeFlag {
		return
	}
	client.writeChan <- nil
	client.closeFlag = true
	client.wg.Wait()
}

func (client *Client) Destroy() {
	client.Lock()
	defer client.Unlock()
	_ = client.conn.Close()
	if !client.closeFlag {
		close(client.writeChan)
		client.closeFlag = true
	}
}
