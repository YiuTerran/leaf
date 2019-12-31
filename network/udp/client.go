package udp

import (
	"net"
	"sync"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/processor"
	"golang.org/x/xerrors"
)

//udp是无连接的，直接将字节流读入读出即可
//操作系统将客户端一次发过来的包一次性放入缓冲区，因此没有所谓的粘包问题；go使用了linux的UDP设计实现，bind本质上只是绑定了一个地址
//但是单次包最大长度是有限制的(理论最大值：65507)，如果超出这个限制，就要拆包；但是由于UDP不可靠，这种情况下就要自己实现TCP的很多功能
//另外，由于MTU的限制，UDP的长度最好在512字节之内（参考http://dwz.win/vN5)
//所以UDP这里的设计直接把processor嵌入进去了，每次发送和接收都是独立的，不过设置了发送缓冲区

const (
	MaxPacketSize     = 65507
	DefaultPacketSize = 1024
)

var (
	ChanFullError = xerrors.New("write chan full")
	InitError     = xerrors.New("fail to init")
)

type Client struct {
	sync.Mutex
	Addr       string
	BufferSize int
	MaxTry     int //最多尝试次数
	Processor  processor.Processor

	closeFlag bool
	writeChan chan []byte
	conn      *net.UDPConn
	wg        *sync.WaitGroup
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
	client.writeChan = make(chan []byte, client.BufferSize)
	client.wg = &sync.WaitGroup{}
	if client.Processor == nil {
		log.Fatal("udp client no processor registered!")
	}

	client.conn = client.dial()
	if client.conn == nil {
		return InitError
	}
	client.wg.Add(1)
	go func() {
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
	}()
	return nil
}

func (client *Client) dial() *net.UDPConn {
	rAddr, err := net.ResolveUDPAddr("udp", client.Addr)
	if err != nil {
		log.Error("fail to resolve add for udp:%v", client.Addr)
		return nil
	}
	conn, err := net.DialUDP("udp", nil, rAddr)
	if err == nil || client.closeFlag {
		return conn
	}

	log.Error("connect to %v error: %v", client.Addr, err)
	return nil
}

func (client *Client) WriteMsg(msg interface{}) error {
	args, err := client.Processor.Marshal(msg)
	if err != nil {
		return err
	}
	client.Lock()
	defer client.Unlock()
	if client.closeFlag {
		return nil
	}
	if len(client.writeChan) == cap(client.writeChan) {
		return ChanFullError
	}
	l := 0
	for i := 0; i < len(args); i++ {
		l += len(args[i])
	}
	bytes := make([]byte, l)
	l = 0
	for i := 0; i < len(args); i++ {
		copy(bytes[l:], args[i])
		l += len(args[i])
	}
	client.writeChan <- bytes
	return nil
}

func (client *Client) ReadMsg() (interface{}, error) {
	buffer := make([]byte, DefaultPacketSize)
	n, err := client.conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	buffer = buffer[:n]
	msg, err := client.Processor.Unmarshal(buffer)
	return msg, err
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
