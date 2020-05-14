package gate

import (
	"time"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/ws"
	"github.com/YiuTerran/leaf/processor"
)

type WsClient struct {
	Server        string
	HttpTimeout   time.Duration
	MsgProcessor  processor.Processor
	RPCServer     *chanrpc.Server
	AutoReconnect bool
}

func (w *WsClient) Processor() processor.Processor {
	return w.MsgProcessor
}

func (w *WsClient) AgentChanRPC() *chanrpc.Server {
	return w.RPCServer
}

func (w *WsClient) Run(closeSig chan struct{}) {
	var wsClient *ws.Client
	if w.Server != "" {
		wsClient = &ws.Client{
			Addr:             w.Server,
			HandshakeTimeout: w.HttpTimeout,
			AutoReconnect:    w.AutoReconnect,
			NewAgent: func(conn *ws.Conn) network.Agent {
				a := &agent{conn: conn, gate: w}
				if w.RPCServer != nil {
					w.RPCServer.Go(AgentCreatedEvent, a)
				}
				return a
			},
		}
	}
	if wsClient != nil {
		wsClient.Start()
	}
	<-closeSig
	if wsClient != nil {
		wsClient.Close()
	}
}

func (w *WsClient) OnDestroy() {}
