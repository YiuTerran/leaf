package module

import (
	"runtime"
	"sync"

	"github.com/YiuTerran/leaf/log"
)

//leaf的模块，启动时按注册顺序逐个启动
//模块的版本用于热加载，当程序调用Reload时，leaf重新获取当前需要激活的mod
//然后和目前正在运行的mod进行比对，如果name不存在则进行激活；
//如果name存在但是version不一致，需要关闭旧的mod，然后再激活新的
type Module interface {
	Name() string
	Version() int
	OnInit()
	OnDestroy()
	Run(closeSig chan struct{})
}

type module struct {
	mi       Module
	closeSig chan struct{}
	wg       sync.WaitGroup
}

var (
	mods = make(map[string]*module)
	lock sync.RWMutex
)

func Register(mis ...Module) {
	lock.Lock()
	defer lock.Unlock()
	for _, mi := range mis {
		if old, ok := mods[mi.Name()]; ok {
			if mi.Version() != old.mi.Version() {
				log.Warn("module %s version changed, restart", mi.Name())
				destroyMod(old)
			} else {
				log.Debug("ignore module %s register", mi.Name())
				return
			}
		}
		m := new(module)
		m.mi = mi
		m.closeSig = make(chan struct{}, 1)
		mods[mi.Name()] = m
		mi.OnInit()
		m.wg.Add(1)
		go run(m)
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
}

//由于修改了执行逻辑，Destroy这里变得无序，在leaf那边注册一个before close的回调来
func Destroy() {
	lock.RLock()
	defer lock.RUnlock()
	for _, mod := range mods {
		destroyMod(mod)
	}
}

func run(m *module) {
	m.mi.Run(m.closeSig)
	m.wg.Done()
}
