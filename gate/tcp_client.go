package gate

import (
	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/tcp"
	"github.com/YiuTerran/leaf/processor"
)

type TcpClient struct {
	Server        string
	MsgProcessor  processor.Processor
	RPCServer     *chanrpc.Server
	BinaryParser  tcp.IParser
	AutoReconnect bool
}

func (c *TcpClient) Processor() processor.Processor {
	return c.MsgProcessor
}

func (c *TcpClient) AgentChanRPC() *chanrpc.Server {
	return c.RPCServer
}

func (c *TcpClient) Run(closeSig chan struct{}) {
	var tcpClient *tcp.Client
	if c.Server != "" {
		tcpClient = &tcp.Client{
			Addr:          c.Server,
			AutoReconnect: c.AutoReconnect,
			Parser:        c.BinaryParser,
			NewAgent: func(conn *tcp.Conn) network.Agent {
				a := &agent{conn: conn, gate: c}
				if c.RPCServer != nil {
					c.RPCServer.Go(AgentCreatedEvent, a)
				}
				return a
			},
		}
	}
	if tcpClient != nil {
		tcpClient.Start()
	}
	<-closeSig
	if tcpClient != nil {
		tcpClient.Close()
	}
}

func (c *TcpClient) OnDestroy() {}
