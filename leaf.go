package leaf

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/YiuTerran/leaf/console"
	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/module"
)

var (
	closeChannel  = make(chan os.Signal, 1)
	reloadChannel = make(chan os.Signal, 1)
)

//手动关闭服务
func CloseServer() {
	closeChannel <- os.Interrupt
}

//重启服务
func RestartServer() {
	closeChannel <- syscall.SIGUSR1
}

//内部热加载
func ReloadServer() {
	reloadChannel <- syscall.SIGUSR2
}

//热加载
func reload(getMods func() []module.Module) {
	for {
		sig := <-reloadChannel
		if sig != syscall.SIGUSR2 {
			return
		}
		module.Register(getMods()...)
	}
}

func Run(consolePort int, getMods func() []module.Module, beforeClose func()) {
	//注意在此之前要调用log.InitLogger
	log.Info("Leaf %v starting up", version)
	// module
	module.Register(getMods()...)
	// console
	console.Init(consolePort)
	//注册热加载信号
	signal.Notify(reloadChannel, syscall.SIGUSR2)
	go reload(getMods)
	//关闭&&重启
	signal.Notify(closeChannel, os.Interrupt, os.Kill, syscall.SIGUSR1)
	sig := <-closeChannel
	reloadChannel <- sig
	if beforeClose != nil {
		beforeClose()
	}
	//清理资源，由于加入了热重启功能，很多地方的代码都要修改，尤其是使用了全局变量的地方
	signal.Stop(reloadChannel)
	signal.Stop(closeChannel)
	console.Destroy()
	module.Destroy()
	if sig == syscall.SIGUSR1 {
		log.Info("Leaf hot restarting...")
		Run(consolePort, getMods, beforeClose)
	} else {
		log.Info("Leaf closing down (signal: %v)", sig)
	}
}
