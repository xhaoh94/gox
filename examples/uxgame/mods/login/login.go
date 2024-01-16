package login

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
	"github.com/xhaoh94/gox/examples/uxgame/game"
)

type (
	//LoginModule Gate模块
	LoginModule struct {
		gox.Module
	}
)

// OnInit 初始化
func (m *LoginModule) OnInit() {
	protoreg.RegisterRpcCmd(pb.CMD_C2S_LoginGame, m.LoginGame)
}

func (m *LoginModule) OnStart() {

}

func (m *LoginModule) LoginGame(ctx context.Context, session types.ISession, req *pb.C2S_LoginGame) (*pb.S2C_LoginGame, error) {

	cfgs := gox.NetWork.GetServiceEntitys(types.WithType(game.Gate)) //获取Gate服务器配置
	if len(cfgs) == 0 {
		logger.Error().Msgf("没获取到[%s]对应的服务器配置", game.Gate)
		return &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown}, nil
	}
	gateCfg := cfgs[0]
	logger.Info().Msgf("[Rpcaddr:%s]", gateCfg.GetRpcAddr())
	conn := gox.NetWork.Rpc().GetClientConnByAddr(gateCfg.GetRpcAddr()) //创建session连接Gate服务器
	loginSession := pb.NewILoginGameClient(conn)
	resp, err := loginSession.LoginGame(ctx, req) //向Gate服务器请求token
	if err != nil {
		logger.Debug().Err(err)
		return &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown}, nil
	}
	return resp, nil //结果返回客户端
}
