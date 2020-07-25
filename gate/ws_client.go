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
	MsgTextFormat bool
	HttpTimeout   time.Duration
	MsgProcessor  processor.Processor
	RPCServer     *chanrpc.Server
	AutoReconnect bool
	UserData      interface{}
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
			ConnNum:          1,
			ConnectInterval:  3 * time.Second,
			HandshakeTimeout: w.HttpTimeout,
			AutoReconnect:    w.AutoReconnect,
			TextFormat:       w.MsgTextFormat,
			NewAgent: func(conn *ws.Conn) network.Agent {
				a := &agent{conn: conn, gate: w}
				if w.RPCServer != nil {
					w.RPCServer.Go(AgentCreatedEvent, a, w.UserData)
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
