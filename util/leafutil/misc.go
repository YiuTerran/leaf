package leafutil

import (
	"fmt"
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
	ConfigDir   = "_config"
	ResourceDir = "resource"
)

//可以在执行go run的时候直接把当前目录当参数传进来，然后在此基础上查找
func FindPathFrom(root string, name string) string {
	dir := root
	prev := ""
	if root == "" {
		return FindPath(name)
	}
	x := filepath.Join(dir, name)
	for !fs.Exists(x) {
		if dir == prev {
			if log.IsInit() {
				log.Error("can't find path from %s, it should be named `%s`", root, name)
			} else {
				panic(fmt.Sprintf("path %s can't be found from %s!!\n", name, root))
			}
			return ""
		}
		prev = dir
		dir = filepath.Dir(dir)
		x = filepath.Join(dir, name)
	}
	return x
}

//从程序所在目录逐层往上找直到找到
//注意：go run的执行文件在临时目录，一般是(/var/folders/)，这种方法是找不到的
func FindPath(name string) string {
	dir, _ := os.Executable()
	root := dir
	return FindPathFrom(root, name)
}

//约定程序二进制文件和_config文件夹在同一层

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

func NewProtoWsGate(wsAddr string, chanRPC *chanrpc.Server) *gate.WsGate {
	return &gate.WsGate{
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
