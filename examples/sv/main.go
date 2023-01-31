package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/app"

	"github.com/xhaoh94/gox/engine/helper/codechelper"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/network/service/kcp"
	"github.com/xhaoh94/gox/engine/network/service/ws"
	"github.com/xhaoh94/gox/examples/sv/game"
	"github.com/xhaoh94/gox/examples/sv/mods"
)

func main() {

	var sid uint
	flag.UintVar(&sid, "sid", uint(strhelper.StringToHash(commonhelper.GetUUID())), "uuid")
	var sType, iAddr, oAddr string
	flag.StringVar(&sType, "type", "all", "服务类型")
	flag.StringVar(&iAddr, "iAddr", "127.0.0.1:10001", "服务地址")
	flag.StringVar(&oAddr, "oAddr", "127.0.0.1:10002", "服务地址")
	// rAddr := *flag.String("grpcAddr", "127.0.0.1:10003", "grpc服务地址")
	flag.Parse()
	app.LoadAppConfig("gox.yaml")
	engine := gox.NewEngine(sid, sType, "1.0.0")
	game.Engine = engine
	engine.SetModule(new(mods.MainModule))
	// engine.SetCodec()
	engine.SetInteriorService(new(kcp.KService), iAddr, codechelper.Json)
	engine.SetOutsideService(new(ws.WService), oAddr, codechelper.Json)
	// engine.SetGrpcAddr(rAddr)
	engine.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	engine.Shutdown()
	os.Exit(1)
}
