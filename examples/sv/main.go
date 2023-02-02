package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/xhaoh94/gox"

	"github.com/xhaoh94/gox/engine/helper/codechelper"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/network/service/kcp"
	"github.com/xhaoh94/gox/engine/network/service/ws"
	"github.com/xhaoh94/gox/examples/sv/game"
	"github.com/xhaoh94/gox/examples/sv/mods"
)

func main() {

	// var sid uint
	// flag.UintVar(&sid, "sid", uint(strhelper.StringToHash(commonhelper.GetUUID())), "uuid")
	// var sType, iAddr, oAddr string
	// flag.StringVar(&sType, "type", "all", "服务类型")
	// flag.StringVar(&iAddr, "iAddr", "127.0.0.1:10001", "服务地址")
	// flag.StringVar(&oAddr, "oAddr", "127.0.0.1:10002", "服务地址")
	appConfPath := *flag.String("appConf", "", "grpc服务地址")
	flag.Parse()
	if appConfPath == "" {
		log.Fatalf("需要启动配置文件路径")
	}

	engine := gox.NewEngine(appConfPath)
	network := network.New(engine)
	network.SetInteriorService(new(kcp.KService), codechelper.Json)
	network.SetOutsideService(new(ws.WService), codechelper.Json)

	engine.SetNetWork(network)
	engine.SetModule(new(mods.MainModule))
	game.Engine = engine
	engine.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	engine.Shutdown()
	os.Exit(1)
}
