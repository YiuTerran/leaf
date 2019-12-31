package gate

import (
	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/processor"
)

const (
	AgentCreatedEvent     = "NewAgent"
	AgentBeforeCloseEvent = "CloseAgent"
)

//路由
type IGate interface {
	Processor() processor.Processor
	AgentChanRPC() *chanrpc.Server
}
