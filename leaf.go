package leaf

import (
	"os"
	"os/signal"

	"github.com/YiuTerran/leaf/console"
	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/module"
)

func Run(consolePort int, mods ...module.Module) {
	//注意在此之前要调用log.InitLogger
	log.Info("Leaf %v starting up", version)

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()
	// console
	console.Init(consolePort)
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Info("Leaf closing down (signal: %v)", sig)
	console.Destroy()
	module.Destroy()
}
