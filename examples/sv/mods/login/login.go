package login

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/types"
)

type (
	//LoginModule 主模块
	LoginModule struct {
		gox.Module
	}
)

//OnStart 初始化
func (mm *LoginModule) OnStart() {
	mm.Register(netpack.C2S_TEST, mm.test)
	mm.RegisterRPC(mm.test2)
}

func (m *LoginModule) test(ctx context.Context, session types.ISession, req *netpack.ReqTest) {
	xlog.Info("服务器>>>接受的消息%v", req)
	session.Send(netpack.S2S_TEST, &netpack.RspTest{Token: "这是普通的响应"})
	// rsp := &netpack.RspTest{}
}

func (m *LoginModule) test2(ctx context.Context, req *netpack.ReqTest) *netpack.RspTest {
	xlog.Info("服务器>>>接受RPC消息%v", req)
	return &netpack.RspTest{Token: "这是rpc响应"}
}
