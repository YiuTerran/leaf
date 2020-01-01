package udp

import (
	"net"
	"time"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/processor"
	"github.com/YiuTerran/leaf/util/netutil"
)

//无连接的一次性的client，用于同步处理
type Client struct {
	serverAddr *net.UDPAddr
	processor  processor.Processor
	conn       *net.UDPConn
}

func NewClient(addr string, processor processor.Processor) *Client {
	rAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Error("fail to resolve udp addr %s", addr)
		return nil
	}
	var conn *net.UDPConn
	if conn, err = net.ListenUDP("udp", nil); err != nil {
		log.Error("fail to listen udp")
		return nil
	}
	return &Client{
		serverAddr: rAddr,
		processor:  processor,
		conn:       conn,
	}
}

//同步请求并等待响应调用回调
func (c *Client) Request(msg interface{}, callback func(interface{}), timeout time.Duration) error {
	bs, err := c.processor.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err = c.conn.WriteToUDP(netutil.MergeBytes(bs), c.serverAddr); err != nil {
		return err
	}
	if callback != nil {
		if timeout > 0 {
			_ = c.conn.SetDeadline(time.Now().Add(timeout))
		}
		buffer := make([]byte, DefaultPacketSize)
		n, _, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		m, err := c.processor.Unmarshal(buffer[:n])
		if err != nil {
			return err
		}
		callback(m)
	}
	return nil
}

//关闭客户端
func (c *Client) Close() {
	_ = c.conn.Close()
}
