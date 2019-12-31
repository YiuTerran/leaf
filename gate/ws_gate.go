package gate

import (
	"time"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/ws"
	"github.com/YiuTerran/leaf/processor"
)

type WebsocketGate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	MsgProcessor    processor.Processor
	RPCServer       *chanrpc.Server

	Addr        string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string
}

func (gate *WebsocketGate) Processor() processor.Processor {
	return gate.MsgProcessor
}

func (gate *WebsocketGate) AgentChanRPC() *chanrpc.Server {
	return gate.RPCServer
}

func (gate *WebsocketGate) Run(closeSig chan struct{}) {
	var wsServer *ws.Server
	if gate.Addr != "" {
		wsServer = new(ws.Server)
		wsServer.Addr = gate.Addr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = func(conn *ws.Conn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.RPCServer != nil {
				gate.RPCServer.Go(AgentCreatedEvent, a)
			}
			return a
		}
	}
	if wsServer != nil {
		wsServer.Start()
	}
	<-closeSig
	if wsServer != nil {
		wsServer.Close()
	}
}

func (gate *WebsocketGate) OnDestroy() {}
