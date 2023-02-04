package main

import (
	"flag"
	"log"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/codechelper"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/network/service/kcp"
	"github.com/xhaoh94/gox/engine/network/service/ws"
	"github.com/xhaoh94/gox/examples/sv/mods"
)

func main() {

	// var sid uint
	// flag.UintVar(&sid, "sid", uint(strhelper.StringToHash(commonhelper.GetUUID())), "uuid")
	// var sType, iAddr, oAddr string
	// flag.StringVar(&sType, "type", "all", "服务类型")
	// flag.StringVar(&iAddr, "iAddr", "127.0.0.1:10001", "服务地址")
	// flag.StringVar(&oAddr, "oAddr", "127.0.0.1:10002", "服务地址")

	var appConfPath string
	flag.StringVar(&appConfPath, "appConf", "app_1.yaml", "启动配置")
	flag.Parse()
	if appConfPath == "" {
		log.Fatalf("需要启动配置文件路径")
	}
	gox.Init(appConfPath)
	network := network.New()
	network.SetInteriorService(new(kcp.KService), codechelper.Json)
	network.SetOutsideService(new(ws.WService), codechelper.Json)

	gox.SetNetWork(network)
	gox.SetModule(new(mods.MainModule))
	gox.Run()

}
