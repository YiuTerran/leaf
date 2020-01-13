package module

import (
	"runtime"
	"sync"

	"github.com/YiuTerran/leaf/chanrpc"
	"github.com/YiuTerran/leaf/log"
)

//leaf的模块，启动时按注册顺序逐个启动
//模块的版本用于热加载，当程序调用Reload时，leaf重新获取当前需要激活的mod，根据Action执行对应的操作

type Action int

const (
	Nothing Action = iota
	New
	Update
	Delete
)

type Module interface {
	Name() string
	Version() string //用于标示module配置是否修改，可以用一些关键的配置hash
	OnInit()
	OnDestroy()
	Run(closeSig chan struct{})
	RPCServer() *chanrpc.Server //如果module可以接受外来的指令，则必须有一个chanrpc server，否则返回nil即可
}

type module struct {
	mi       Module
	closeSig chan struct{}
	wg       sync.WaitGroup
}

var (
	mods = make(map[string]*module)
	lock sync.Mutex
)

func Reload(actionMds map[Action][]Module) {
	lock.Lock()
	defer lock.Unlock()
	for action, mis := range actionMds {
		//不管是哪种行为，都要删除旧模块
		for _, mi := range mis {
			if old, ok := mods[mi.Name()]; !ok {
				if action != New {
					log.Info("no active module %s, ignore", mi.Name())
				}
			} else {
				destroyMod(old)
				if action == New {
					log.Warn("register new module but old exists, destroy module %s", mi.Name())
				}
			}
		}
		//新增模块
		if action == New || action == Update {
			for _, mi := range mis {
				m := new(module)
				m.mi = mi
				m.closeSig = make(chan struct{}, 1)
				mods[mi.Name()] = m
				mi.OnInit()
				m.wg.Add(1)
				go run(m)
				log.Info("module %s registered", mi.Name())
			}
		}
	}
}

func destroyMod(mod *module) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, log.LenStackBuf)
			l := runtime.Stack(buf, false)
			log.Error("panic when destroy module %s, %v: %s", mod.mi.Name(), r, buf[:l])
		}
	}()
	mod.closeSig <- struct{}{}
	mod.wg.Wait()
	mod.mi.OnDestroy()
	delete(mods, mod.mi.Name())
	log.Info("module %s destroyed", mod.mi.Name())
}

//由于修改了执行逻辑，Destroy这里变得无序，在leaf那边注册一个before close的回调来
func Destroy() {
	lock.Lock()
	defer lock.Unlock()
	for _, mod := range mods {
		destroyMod(mod)
	}
}

func run(m *module) {
	m.mi.Run(m.closeSig)
	m.wg.Done()
}
