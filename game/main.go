package main

import (
	"flag"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/network/service/kcp"
	"github.com/xhaoh94/gox/engine/network/service/tcp"
	"github.com/xhaoh94/gox/game/mods"
	"github.com/xhaoh94/gox/util"
)

func main() {
	flag.StringVar(&app.SID, "sid", util.GetUUID(), "uuid")
	flag.StringVar(&app.ServiceType, "type", "all", "服务类型")
	iAddr := flag.String("iAddr", "127.0.0.1:10001", "服务地址")
	oAddr := flag.String("oAddr", "127.0.0.1:10002", "服务地址")
	rAddr := flag.String("grpcAddr", "127.0.0.1:10003", "grpc服务地址")
	flag.Parse()
	gox.SetCodec(new(codec.ProtobufCodec))
	gox.SetInteriorService(new(tcp.TService), *iAddr)
	gox.SetOutsideService(new(kcp.KService), *oAddr)
	gox.SetGrpcAddr(*rAddr)
	gox.SetModule(new(mods.MainModule))
	gox.Start("xhgo.ini")
	gox.Shutdown()
}
