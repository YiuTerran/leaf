package leaf

import (
	"os"
	"os/signal"

	"github.com/YiuTerran/leaf/console"
	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/module"
)

var (
	closeChannel    = make(chan os.Signal, 1)
	internalChannel = make(chan int, 1)

	quitSig   = 1
	reloadSig = 2
)

type GetModules func() map[module.Action][]module.Module

//手动关闭服务
func CloseServer() {
	closeChannel <- os.Kill
}

//内部热加载
func ReloadServer() {
	internalChannel <- reloadSig
}

//热加载
func reload(getMods GetModules) {
	for {
		sig := <-internalChannel
		if sig == quitSig {
			break
		} else if sig == reloadSig {
			module.Reload(getMods())
		}
	}
}

//一般运行模式：开启模块热加载特性
func Run(consolePort int, getMods GetModules, beforeClose func()) {
	//注意在此之前要调用log.InitLogger
	log.Info("Server %v starting up", version)
	// module
	module.Reload(getMods())
	// console
	console.Init(consolePort)
	//注册热加载信号
	go reload(getMods)
	//关闭&&重启
	signal.Notify(closeChannel, os.Interrupt, os.Kill)
	<-closeChannel
	internalChannel <- quitSig
	if beforeClose != nil {
		beforeClose()
	}
	signal.Stop(closeChannel)
	console.Destroy()
	module.Destroy()
	log.Info("Server closing down")
}

//模块以静态模式加载（关闭热加载特性）
func StaticRun(consolePort int, mods []module.Module) {
	//注意在此之前要调用log.InitLogger
	log.Info("Server %v starting up", version)
	module.StaticLoad(mods)
	console.Init(consolePort)
	signal.Notify(closeChannel, os.Interrupt, os.Kill)
	<-closeChannel
	console.Destroy()
	module.Destroy()
	log.Info("Server closing down")
}
