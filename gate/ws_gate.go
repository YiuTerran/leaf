package gate

import (
	"net/http"
	"time"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/ws"
	"github.com/YiuTerran/leaf/processor"
)

type WsGate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	MsgProcessor    processor.Processor
	MsgTextFormat   bool
	AuthFunc        func(*http.Request) (bool, interface{})
	RPCServer       *chanrpc.Server

	Addr        string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string
}

func (gate *WsGate) Processor() processor.Processor {
	return gate.MsgProcessor
}

func (gate *WsGate) AgentChanRPC() *chanrpc.Server {
	return gate.RPCServer
}

func (gate *WsGate) Run(closeSig chan struct{}) {
	var wsServer *ws.Server
	if gate.Addr != "" {
		wsServer = new(ws.Server)
		wsServer.Addr = gate.Addr
		wsServer.TextFormat = gate.MsgTextFormat
		wsServer.AuthFunc = gate.AuthFunc
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = func(conn *ws.Conn) network.Agent {
			a := &agent{conn: conn, gate: gate, userData: conn.UserData()}
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

func (gate *WsGate) OnDestroy() {}
