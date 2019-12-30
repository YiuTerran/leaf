package gate

import (
	"time"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/network"
	"github.com/YiuTerran/leaf/network/tcp"
	"github.com/YiuTerran/leaf/network/ws"
	"github.com/YiuTerran/leaf/processor"
)

const (
	AgentCreatedEvent     = "NewAgent"
	AgentBeforeCloseEvent = "CloseAgent"
)

type Gate struct {
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       uint32
	Processor       processor.Processor
	AgentChanRPC    *chanrpc.Server

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string
	LenMsgLen    int
	LittleEndian bool
}

func (gate *Gate) Run(closeSig chan struct{}) {
	var wsServer *ws.Server
	if gate.WSAddr != "" {
		wsServer = new(ws.Server)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.CertFile = gate.CertFile
		wsServer.KeyFile = gate.KeyFile
		wsServer.NewAgent = func(conn *ws.Conn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go(AgentCreatedEvent, a)
			}
			return a
		}
	}

	var tcpServer *tcp.Server
	if gate.TCPAddr != "" {
		tcpServer = new(tcp.Server)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum
		parser := tcp.NewDefaultMsgParser()
		parser.SetByteOrder(gate.LittleEndian)
		parser.SetMsgLen(gate.LenMsgLen,0,gate.MaxMsgLen)
		tcpServer.Parser = parser
		tcpServer.NewAgent = func(conn *tcp.Conn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			if gate.AgentChanRPC != nil {
				gate.AgentChanRPC.Go(AgentCreatedEvent, a)
			}
			return a
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}
