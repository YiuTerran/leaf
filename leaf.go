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
	c = make(chan os.Signal, 1)
)

//手动关闭服务
func CloseServer() {
	c <- os.Interrupt
}

//重启服务
func RestartServer() {
	c <- syscall.SIGUSR1
}

func Run(consolePort int, getMods func() []module.Module) {
	//注意在此之前要调用log.InitLogger
	log.Info("Leaf %v starting up", version)
	mods := getMods()
	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()
	// console
	console.Init(consolePort)
	// close
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	console.Destroy()
	module.Destroy()
	if sig == syscall.SIGUSR1 {
		log.Info("Leaf hot restarting...")
		Run(consolePort, getMods)
	} else {
		log.Info("Leaf closing down (signal: %v)", sig)
	}
}
