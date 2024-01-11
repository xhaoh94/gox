package login

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
	"github.com/xhaoh94/gox/examples/sv/game"
)

type (
	//LoginModule Gate模块
	LoginModule struct {
		gox.Module
	}
)

// OnInit 初始化
func (m *LoginModule) OnInit() {
	protoreg.Register(pb.CMD_C2S_LoginGame, m.LoginGame)
}

func (m *LoginModule) OnStart() {

}

func (m *LoginModule) LoginGame(ctx context.Context, session types.ISession, req *pb.C2S_LoginGame) {

	cfgs := gox.NetWork.GetServiceEntitys(types.WithType(game.Gate)) //获取Gate服务器配置
	gateCfg := cfgs[0]
	logger.Info().Msgf("[Rpcaddr:%s]", gateCfg.GetRpcAddr())
	conn := gox.NetWork.Rpc().GetClientConnByAddr(gateCfg.GetRpcAddr()) //创建session连接Gate服务器
	loginSession := pb.NewILoginGameClient(conn)
	resp, err := loginSession.LoginGame(ctx, req) //向Gate服务器请求token
	if err != nil {
		session.Send(pb.CMD_S2C_LoginGame, &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown})
		return
	}
	session.Send(pb.CMD_S2C_LoginGame, resp) //结果返回客户端
}
