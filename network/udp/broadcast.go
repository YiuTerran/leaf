package udp

import (
	"net"
	"time"
)

//广播客户端
//广播的服务端其实就是普通的服务端
type BroadcastClient struct {
	Target string
	Port   int
}

//一个同步的广播
func (bcc *BroadcastClient) Broad(msg []byte, callback func([]byte, net.Addr), timeout time.Duration) error {
	src := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dst := &net.UDPAddr{IP: net.ParseIP(bcc.Target), Port: bcc.Port}
	conn, err := net.ListenUDP("udp", src)
	if err != nil {
		return err
	}
	defer conn.Close()
	n, err := conn.WriteToUDP(msg, dst)
	if err != nil {
		return err
	}
	if callback == nil {
		return nil
	}
	var addr net.Addr
	data := make([]byte, DefaultPacketSize)
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	for {
		if n, addr, err = conn.ReadFrom(data); err == nil {
			callback(data[:n], addr)
		} else {
			break
		}
	}
	return nil
}
