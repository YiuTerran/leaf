package gate

import (
	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/tcp"
	"github.com/YiuTerran/leaf/processor"
)

type TcpGate struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	MsgProcessor    processor.Processor
	RPCServer       *chanrpc.Server
	BinaryParser    tcp.IParser
}

func (gate *TcpGate) Processor() processor.Processor {
	return gate.MsgProcessor
}

func (gate *TcpGate) AgentChanRPC() *chanrpc.Server {
	return gate.RPCServer
}

func (gate *TcpGate) Run(closeSig chan struct{}) {
	var tcpServer *tcp.Server
	if gate.Addr != "" {
		tcpServer = new(tcp.Server)
		tcpServer.Addr = gate.Addr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		tcpServer.Parser = gate.BinaryParser
		tcpServer.NewAgent = func(conn *tcp.Conn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.RPCServer != nil {
				gate.RPCServer.Go(AgentCreatedEvent, a)
			}
			return a
		}
	}

	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *TcpGate) OnDestroy() {}
