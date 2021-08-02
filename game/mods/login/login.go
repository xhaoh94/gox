package login

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/game/netpack"
)

type (
	//LoginModule 主模块
	LoginModule struct {
		gox.Module
	}
)

//OnInit 初始化
func (mm *LoginModule) OnInit() {
	mm.Register(111, mm.test)
	mm.RegisterRPC(mm.test2)
}

func (m *LoginModule) test(ctx context.Context, session types.ISession, req *netpack.ReqTest) {
	xlog.Info("接受的消息%v", req)
	session.Send(2222, &netpack.RspTest{Token: "test"})
	rsp := &netpack.RspTest{}
	session.Call(&netpack.ReqTest{Acc: "xx", Pwd: "cc"}, rsp).Await()
}

func (m *LoginModule) test2(ctx context.Context, req *netpack.ReqTest) *netpack.RspTest {
	xlog.Info("接受的消息2%v", req)
	return &netpack.RspTest{Token: "500"}
}
