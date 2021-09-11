package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/network/sv/kcp"
	"github.com/xhaoh94/gox/engine/network/sv/ws"
	"github.com/xhaoh94/gox/examples/sv/game"
	"github.com/xhaoh94/gox/examples/sv/mods"
	"github.com/xhaoh94/gox/util"
)

func main() {

	var sid uint
	flag.UintVar(&sid, "sid", uint(util.StringToHash(util.GetUUID())), "uuid")
	var sType, iAddr, oAddr string
	flag.StringVar(&sType, "type", "all", "服务类型")
	flag.StringVar(&iAddr, "iAddr", "127.0.0.1:10001", "服务地址")
	flag.StringVar(&oAddr, "oAddr", "127.0.0.1:10002", "服务地址")
	// rAddr := *flag.String("grpcAddr", "127.0.0.1:10003", "grpc服务地址")
	flag.Parse()
	app.LoadAppConfig("gox.ini")
	engine := gox.NewEngine(sid, sType, "1.0.0")
	game.Engine = engine
	engine.SetModule(new(mods.MainModule))
	engine.SetCodec(codec.Protobuf)
	engine.SetInteriorService(new(kcp.KService), iAddr)
	engine.SetOutsideService(new(ws.WService), oAddr)
	// engine.SetGrpcAddr(rAddr)
	engine.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	engine.Shutdown()
	os.Exit(1)
}
