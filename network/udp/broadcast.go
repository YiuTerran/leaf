package udp

import (
	"net"
	"time"
)

//广播客户端
type Broadcast struct {
	Target string
	Port   int
}

//一个同步的广播
func (broadcast *Broadcast) Broad(msg []byte, callback func([]byte, net.Addr), timeout time.Duration) error {
	src := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dst := &net.UDPAddr{IP: net.ParseIP(broadcast.Target), Port: broadcast.Port}
	conn, err := net.ListenUDP("udp", src)
	if err != nil {
		return err
	}
	defer conn.Close()
	n, err := conn.WriteToUDP(msg, dst)
	if err != nil {
		return err
	}
	var addr net.Addr
	data := make([]byte, DefaultPacketSize)
	if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	n, addr, err = conn.ReadFrom(data)
	if err == nil {
		callback(data[:n], addr)
	} else {
		return err
	}
	return nil
}
