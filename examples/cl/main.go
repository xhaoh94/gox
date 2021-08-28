package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/network/service/tcp"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
	"github.com/xhaoh94/gox/xdef"
)

type (
	//MainModule 主模块
	MainModule struct {
		gox.Module
	}
)

//OnStart 初始化
func (mm *MainModule) OnStart() {
	mm.Register(netpack.S2S_TEST, mm.RspTest)
	mm.GetEngine().GetEvent().On(xdef.START_ENGINE_OK, mm.Init)
}

func (mm *MainModule) Init() {
	session := mm.GetSessionByAddr("127.0.0.1:10002")
	session.Send(netpack.C2S_TEST, &netpack.ReqTest{Acc: "xhaoh94", Pwd: "123456"})
}

func (mm *MainModule) RspTest(ctx context.Context, session types.ISession, rsp *netpack.RspTest) {
	fmt.Printf("客户端收到数据 %v", rsp)
}

//模拟客户端发数据
func main() {
	sid := *flag.Uint("sid", uint(util.StringToHash(util.GetUUID())), "uuid")
	sType := *flag.String("type", "client", "服务类型")
	addr := *flag.String("addr", "127.0.0.1:9999", "服务地址")
	flag.Parse()
	engine := gox.NewEngine(sid, sType, "1.0.0")
	engine.SetModule(new(MainModule))
	engine.SetCodec(new(codec.JsonCodec))
	engine.SetOutsideService(new(tcp.TService), addr)
	engine.Start("")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	engine.Shutdown()
	os.Exit(1)
}
