package udp

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/YiuTerran/leaf/log"
)

//udp是无连接的，直接将字节流读入读出即可
//操作系统将客户端一次发过来的包一次性放入缓冲区，因此没有所谓的粘包问题
//但是单次包最大长度是有限制的(理论最大值：65507)，如果超出这个限制，就要拆包；但是由于UDP不可靠，这种情况下就要自己实现TCP的很多功能
//另外，由于MTU的限制，UDP的长度最好在512字节之内（参考http://dwz.win/vN5)

const (
	MaxPacketSize     = 65507
	DefaultPacketSize = 512
)

type Client struct {
	sync.Mutex
	Addr            string
	PendingWriteNum int
	ConnNum         int
	Timeout         time.Duration
	MaxTry          int //最多尝试次数
	Parser          io.ReadWriter

	closeFlag bool
	wg        sync.WaitGroup
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
	}
	if client.MaxTry <= 0 {
		client.MaxTry = 3
	}
	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 100
	}
	client.closeFlag = false
}

func (client *Client) dial() net.Conn {
	for {
		rAddr, err := net.ResolveUDPAddr("udp", client.Addr)
		if err != nil {
			log.Fatal("fail to resolve add for udp:%v", client.Addr)
		}
		conn, err := net.DialUDP("udp", nil, rAddr)
		if err == nil || client.closeFlag {
			return conn
		}

		log.Error("connect to %v error: %v", client.Addr, err)
		continue
	}
}

func (client *Client) connect(){
	defer
}