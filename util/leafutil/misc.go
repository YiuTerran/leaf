package leafutil

import (
	"os"
	"path/filepath"
	"time"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/gate"
	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/module"
	"github.com/YiuTerran/leaf/processor/protobuf"
	"github.com/YiuTerran/leaf/util/fs"
)

const (
	configDir = "_config"
)

//约定程序二进制文件和_config文件夹在同一层
//否则逐层往上找直到找到（方便进行单元测试）
func GetConfigDir() string {
	wd, _ := os.Getwd()

	x := filepath.Join(wd, configDir)
	for !fs.Exists(x) {
		if wd == "/" {
			log.Fatal("can't find config dir, it should be named `%s`", configDir)
		}
		wd = filepath.Dir(wd)
		x = filepath.Join(wd, configDir)
	}
	return x
}

type TextDuration struct {
	time.Duration
}

func (d *TextDuration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

//通用recover函数，在单独协程的最开始使用defer调用
func RecoverFromPanic(cb func()) {
	if r := recover(); r != nil {
		log.Error("recover from panic!!!, error:%v", r)
		if cb != nil {
			cb()
		}
	}
}

const (
	// server conf
	PendingWriteNum = 2000
	MaxMsgLen       = 1 * 1024 * 1024 // 最大长度为1M
	HTTPTimeout     = 5 * time.Second
	LenMsgLen       = 4
	MaxConnNum      = 20000

	// skeleton conf
	GoLen              = 10000
	TimerDispatcherLen = 10000
	AsyncCallLen       = 10000
	ChanRPCLen         = 10000
)

func NewSkeleton() *module.Skeleton {
	skeleton := &module.Skeleton{
		GoLen:              GoLen,
		TimerDispatcherLen: TimerDispatcherLen,
		AsyncCallLen:       AsyncCallLen,
		ChanRPCServer:      chanrpc.NewServer(ChanRPCLen),
	}
	skeleton.Init()
	return skeleton
}

var (
	ProtoProcessor = protobuf.NewProcessor()
)

func NewProtoWsGate(wsAddr string, chanRPC *chanrpc.Server) *gate.WebsocketGate {
	return &gate.WebsocketGate{
		MaxConnNum:      MaxConnNum,
		PendingWriteNum: PendingWriteNum,
		MaxMsgLen:       MaxMsgLen,
		Addr:            wsAddr,
		HTTPTimeout:     HTTPTimeout,
		RPCServer:       chanRPC,
		MsgProcessor:    ProtoProcessor,
	}
}

func CheckAuth(ag gate.Agent) bool {
	if ag == nil {
		return false
	}
	if ag.UserData() == nil {
		ag.Close()
		return false
	}
	return true
}
