package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/network/sv/kcp"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
)

type (
	//MainModule 主模块
	MainModule struct {
		gox.Module
		session types.ISession
	}
)

func (m *MainModule) OnInit() {
	m.Register(netpack.CMD_G2C_Login, m.RspToken)
	m.Register(netpack.CMD_L2C_Login, m.RspLogin)
	m.Register(netpack.CMD_L2C_Enter, m.RspEnter)

}

//OnStart
func (m *MainModule) OnStart() {
	xlog.Debug("test")
	time.Sleep(1 * time.Second)
	session := m.GetSessionByAddr("127.0.0.1:10002") //向gate服务器请求token
	session.Send(netpack.CMD_C2G_Login, &netpack.C2G_Login{User: "xhaoh94", Password: "123456"})
	m.session = session
}

func (m *MainModule) RspToken(ctx context.Context, session types.ISession, rsp *netpack.G2C_Login) {
	if rsp.Code != 0 { //请求token错误
		return
	}
	defer session.Close()                                                                           //老的session已经没用了，可以关闭掉
	loginSession := m.GetSessionByAddr(rsp.Addr)                                                    //创建session连接login服务器
	loginSession.Send(netpack.CMD_C2L_Login, &netpack.C2L_Login{User: "xhaoh94", Token: rsp.Token}) //向login服务器请求登录
	m.session = loginSession                                                                        //保存新的session
}

func (m *MainModule) RspLogin(ctx context.Context, session types.ISession, rsp *netpack.L2C_Login) {
	xlog.Debug("登录结果返回Code:%d", rsp.Code)
	session.Send(netpack.CMD_C2L_Enter, &netpack.C2L_Enter{SceneId: 1})
}

func (m *MainModule) RspEnter(ctx context.Context, session types.ISession, rsp *netpack.L2C_Enter) {
	xlog.Debug("进入结果返回Code:%d", rsp.Code)
}

//模拟客户端发数据
func main() {
	sid := *flag.Uint("sid", uint(util.StringToHash(util.GetUUID())), "uuid")
	sType := *flag.String("type", "client", "服务类型")
	addr := *flag.String("addr", "127.0.0.1:9999", "服务地址")
	flag.Parse()
	engine := gox.NewEngine(sid, sType, "1.0.0")
	engine.SetModule(new(MainModule))
	engine.SetCodec(codec.Json)
	engine.SetInteriorService(new(kcp.KService), addr)
	engine.Start("")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	engine.Shutdown()
	os.Exit(1)
}
