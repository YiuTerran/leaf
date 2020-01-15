package leaf

import (
	"os"
	"os/signal"

	"github.com/YiuTerran/leaf/console"
	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/module"
)

var (
	closeChannel  = make(chan os.Signal, 1)
	reloadChannel = make(chan *struct{}, 1)
)

type GetModules func() map[module.Action][]module.Module

//手动关闭服务
func CloseServer() {
	closeChannel <- os.Interrupt
}

//内部热加载
func ReloadServer() {
	reloadChannel <- &struct{}{}
}

//热加载
func reload(getMods GetModules) {
	for {
		sig := <-reloadChannel
		if sig == nil {
			break
		}
		module.Reload(getMods())
	}
}

func Run(consolePort int, getMods GetModules, beforeClose func()) {
	//注意在此之前要调用log.InitLogger
	log.Info("Leaf %v starting up", version)
	// module
	module.Reload(getMods())
	// console
	console.Init(consolePort)
	//注册热加载信号
	go reload(getMods)
	//关闭&&重启
	signal.Notify(closeChannel, os.Interrupt, os.Kill)
	sig := <-closeChannel
	reloadChannel <- nil
	if beforeClose != nil {
		beforeClose()
	}
	//清理资源，由于加入了热重启功能，很多地方的代码都要修改，尤其是使用了全局变量的地方
	signal.Stop(closeChannel)
	console.Destroy()
	module.Destroy()
	log.Info("Leaf closing down", sig)
}
