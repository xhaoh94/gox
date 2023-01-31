package gate

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/examples/pb"
	"github.com/xhaoh94/gox/examples/sv/game"
)

type (
	//GateModule Gate模块
	GateModule struct {
		gox.Module
	}
)

// OnInit 初始化
func (m *GateModule) OnInit() {
	m.Register(netpack.CMD_C2G_Login, m.RspLogin)
	m.Register(100, m.Test)
}

func (m *GateModule) OnStart() {
}

func (m *GateModule) Test(ctx context.Context, session types.ISession, msg *pb.A) {
	xlog.Debug("test [%v]", msg)
	session.Send(100, &pb.B{Id: "test", Etype: 1, Position: &pb.Vector3{X: 0, Y: 1, Z: 2}})
}

func (m *GateModule) RspLogin(ctx context.Context, session types.ISession, msg *netpack.C2G_Login) {

	//TODO 验证账号密码是否正确
	cfgs := m.GetServiceConfListByType(game.Login) //获取login服务器配置
	loginCfg := cfgs[0]
	loginSession := m.GetSessionByAddr(loginCfg.GetInteriorAddr()) //创建session连接login服务器
	Rsp_L2G_Login := &netpack.L2G_Login{}
	b := loginSession.Call(&netpack.G2L_Login{User: msg.User}, Rsp_L2G_Login).Await() //向login服务器请求token

	Rsp_G2C_Login := &netpack.G2C_Login{}
	if b {
		Rsp_G2C_Login.Code = 0
		Rsp_G2C_Login.Addr = loginCfg.GetOutsideAddr()
		Rsp_G2C_Login.Token = Rsp_L2G_Login.Token
	} else {
		Rsp_G2C_Login.Code = 1
	}
	session.Send(netpack.CMD_G2C_Login, Rsp_G2C_Login) //结果返回客户端
}
