package gate

import (
	"net"
	"reflect"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/network"
)

type Agent interface {
	WriteMsg(msg interface{})
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
	UserData() interface{}
	SetUserData(data interface{})
}

type agent struct {
	conn     network.Conn
	gate     IGate
	userData interface{}
}

//这里出现真的错误才要断开连接
func (a *agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message error: %v", err)
			break
		}
		if len(data) == 0 {
			continue
		}
		if a.gate.Processor() != nil {
			msg, err := a.gate.Processor().Unmarshal(data)
			if err != nil {
				log.Debug("unmarshal message error: %v", err)
				break
			}
			if msg == nil {
				continue
			}
			err = a.gate.Processor().Route(msg, a)
			if err != nil {
				log.Debug("route message error: %v", err)
				break
			}
		}
	}
}

func (a *agent) OnClose() {
	if a.gate.AgentChanRPC() != nil {
		err := a.gate.AgentChanRPC().Call0(AgentBeforeCloseEvent, a)
		if err != nil {
			log.Warn("chanrpc error: %v", err)
		}
	}
}

func (a *agent) WriteMsg(msg interface{}) {
	if a.gate.Processor() != nil {
		data, err := a.gate.Processor().Marshal(msg)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			log.Error("write message %v error: %v", reflect.TypeOf(msg), err)
		}
	}
}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}

func (a *agent) UserData() interface{} {
	return a.userData
}

func (a *agent) SetUserData(data interface{}) {
	a.userData = data
}
