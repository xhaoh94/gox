package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/network/codec"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/network/service/ws"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
)

type (
	//MainModule 主模块
	MainModule struct {
		gox.Module
		session types.ISession
	}
)

func (m *MainModule) OnInit() {
	protoreg.Register(netpack.CMD_G2C_Login, m.RspToken)
	protoreg.Register(netpack.CMD_L2C_Login, m.RspLogin)
	protoreg.Register(netpack.CMD_L2C_Enter, m.RspEnter)

}

// OnStart
func (m *MainModule) OnStart() {
	xlog.Debug("test")
	time.Sleep(1 * time.Second)
	session := gox.NetWork.GetSessionByAddr("127.0.0.1:10002") //向gate服务器请求token
	session.Send(netpack.CMD_C2G_Login, &netpack.C2G_Login{User: "xhaoh94", Password: "123456"})
	m.session = session
}

func (m *MainModule) RspToken(ctx context.Context, session types.ISession, rsp *netpack.G2C_Login) {
	if rsp.Code != 0 { //请求token错误
		return
	}
	defer session.Close() //老的session已经没用了，可以关闭掉
	xlog.Debug("返回数据:%v", rsp)
	loginSession := gox.NetWork.GetSessionByAddr(rsp.Addr)                                          //创建session连接login服务器
	loginSession.Send(netpack.CMD_C2L_Login, &netpack.C2L_Login{User: "xhaoh94", Token: rsp.Token}) //向login服务器请求登录
	m.session = loginSession                                                                        //保存新的session
}

func (m *MainModule) RspLogin(ctx context.Context, session types.ISession, rsp *netpack.L2C_Login) {
	xlog.Debug("登录结果返回Code:%d", rsp.Code)
	session.Send(netpack.CMD_C2L_Enter, &netpack.C2L_Enter{SceneId: 1, UnitId: 100})
	session.Send(netpack.CMD_C2L_Enter, &netpack.C2L_Enter{SceneId: 1, UnitId: 200})
	session.Send(netpack.CMD_C2L_Enter, &netpack.C2L_Enter{SceneId: 2, UnitId: 300})
}

func (m *MainModule) RspEnter(ctx context.Context, session types.ISession, rsp *netpack.L2C_Enter) {
	xlog.Debug("进入结果返回Code:%d", rsp.Code)
}

// 模拟客户端发数据
func main() {
	var appConfPath string
	flag.StringVar(&appConfPath, "appConf", "app_1.yaml", "启动配置")
	flag.Parse()
	if appConfPath == "" {
		log.Fatalf("需要启动配置文件路径")
	}
	gox.Init(appConfPath)
	network := network.New()
	network.SetInteriorService(new(ws.WService), codec.Json)
	gox.SetNetWork(network)
	gox.SetModule(new(MainModule))
	gox.Run()

}

var KEY []byte = []byte("key_key_")

const (
	H_B_S byte = 0x01
	H_B_R byte = 0x02
	C_S_C byte = 0x03
	RPC_S byte = 0x04
	RPC_R byte = 0x05
)
