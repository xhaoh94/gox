package login

import (
	"context"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
)

type (
	//LoginModule 登录模块
	LoginModule struct {
		gox.Module
		mux        sync.RWMutex
		user2Token map[string]UserToken
	}
	UserToken struct {
		user  string
		token string
		time  time.Time
	}
)

// OnInit 初始化
func (m *LoginModule) OnInit() {
	m.user2Token = make(map[string]UserToken)
	protoreg.RegisterRpc(m.RspToken)
	protoreg.Register(netpack.CMD_C2L_Login, m.RspLogin)
	protoreg.Register(netpack.CMD_C2L_Enter, m.RspEnter)
}

func (m *LoginModule) OnStart() {

}

func (m *LoginModule) RspLogin(ctx context.Context, session types.ISession, req *netpack.C2L_Login) {
	m.mux.RLock()
	ut, ok := m.user2Token[req.User]
	m.mux.RUnlock()
	if !ok {
		session.Send(netpack.CMD_L2C_Login, &netpack.L2C_Login{Code: 1}) //没有找到对应的token
		return
	}
	if ut.token != req.Token {
		session.Send(netpack.CMD_L2C_Login, &netpack.L2C_Login{Code: 2}) //token不一致
		return
	}
	defer func() {
		m.mux.Lock()
		delete(m.user2Token, req.User)
		m.mux.Unlock()
	}()

	t := time.Now().Sub(ut.time)
	if t.Seconds() > 5 {
		session.Send(netpack.CMD_L2C_Login, &netpack.L2C_Login{Code: 3}) //token已过期
		return
	}

	session.Send(netpack.CMD_L2C_Login, &netpack.L2C_Login{Code: 0}) //返回客户端结果
}

func (m *LoginModule) RspToken(ctx context.Context, req *netpack.G2L_Login) (*netpack.L2G_Login, error) {
	token := commonhelper.NewUUID() //创建user对应的token
	xlog.Debug("创建user[%s]对应的token[%s]", req.User, token)
	m.mux.Lock()
	m.user2Token[req.User] = UserToken{user: req.User, token: token, time: time.Now()} //将user、token保存
	m.mux.Unlock()
	return &netpack.L2G_Login{Token: token}, nil
}

func (m *LoginModule) RspEnter(ctx context.Context, session types.ISession, req *netpack.C2L_Enter) {
	sId := req.SceneId
	backRsp := &netpack.S2L_Enter{}
	err := gox.Location.Call(uint32(sId), &netpack.L2S_Enter{UnitId: req.UnitId}, backRsp).Await() //Actor玩家进入场景

	enterRsp := &netpack.L2C_Enter{}
	if err == nil { //玩家进入场景成功
		rsp := &netpack.S2L_SayHello{}
		xlog.Debug("玩家发言")
		err = gox.Location.Call(uint32(req.UnitId), &netpack.L2S_SayHello{Txt: "你好啊，我是机器人:" + strhelper.ValToString(req.UnitId)}, rsp).Await() //Actor 玩家发言
		if err == nil {
			xlog.Debug("发言返回:%s", rsp.BackTxt)
			enterRsp.Code = 0
		} else {
			enterRsp.Code = 2
		}
	} else {
		enterRsp.Code = 1
		xlog.Error("进入场景错误,%v", err)
	}
	session.Send(netpack.CMD_L2C_Enter, enterRsp)
}
