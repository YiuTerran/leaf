package module

import (
	"runtime"
	"sync"

	"github.com/YiuTerran/leaf/log"
)

//leaf的模块，启动时按注册顺序逐个启动
type Module interface {
	OnInit()
	OnDestroy()
	Run(closeSig chan struct{})
}

type module struct {
	mi       Module
	closeSig chan struct{}
	wg       sync.WaitGroup
}

var mods []*module

func Register(mi Module) {
	m := new(module)
	m.mi = mi
	m.closeSig = make(chan struct{}, 1)

	mods = append(mods, m)
}

func Init() {
	for i := 0; i < len(mods); i++ {
		mods[i].mi.OnInit()
	}

	for i := 0; i < len(mods); i++ {
		m := mods[i]
		m.wg.Add(1)
		go run(m)
	}
}

func Destroy() {
	for i := len(mods) - 1; i >= 0; i-- {
		m := mods[i]
		m.closeSig <- struct{}{}
		m.wg.Wait()
		destroy(m)
	}
}

func run(m *module) {
	m.mi.Run(m.closeSig)
	m.wg.Done()
}

func destroy(m *module) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, log.LenStackBuf)
			l := runtime.Stack(buf, false)
			log.Error("%v: %s", r, buf[:l])
		}
	}()

	m.mi.OnDestroy()
}
