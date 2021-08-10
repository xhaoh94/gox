package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/network/service/kcp"
	"github.com/xhaoh94/gox/engine/network/service/tcp"
	"github.com/xhaoh94/gox/examples/sv/mods"
	"github.com/xhaoh94/gox/util"
)

func main() {
	sid := *flag.Uint("sid", uint(util.StringToHash(util.GetUUID())), "uuid")
	sType := *flag.String("type", "all", "服务类型")
	iAddr := *flag.String("iAddr", "127.0.0.1:10001", "服务地址")
	oAddr := *flag.String("oAddr", "127.0.0.1:10002", "服务地址")
	rAddr := *flag.String("grpcAddr", "127.0.0.1:10003", "grpc服务地址")
	flag.Parse()
	engine := gox.NewEngine(sid, sType, "1.0.0", new(mods.MainModule))
	engine.SetCodec(new(codec.ProtobufCodec))
	engine.SetInteriorService(new(tcp.TService), iAddr)
	engine.SetOutsideService(new(kcp.KService), oAddr)
	engine.SetGrpcAddr(rAddr)
	engine.Start("xhgo.ini")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	engine.Shutdown()
	os.Exit(1)
}
