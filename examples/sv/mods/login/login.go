package login

import (
	"context"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
)

type (
	//LoginModule 主模块
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

//OnStart 初始化
func (m *LoginModule) OnStart() {
	m.user2Token = make(map[string]UserToken)
	m.RegisterRPC(m.RspToken)
	m.Register(netpack.CMD_C2L_Login, m.RspLogin)
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
	t := time.Now().Sub(ut.time)
	if t.Seconds() > 5 {
		session.Send(netpack.CMD_L2C_Login, &netpack.L2C_Login{Code: 3}) //token已过期
		return
	}
	session.Send(netpack.CMD_L2C_Login, &netpack.L2C_Login{Code: 0}) //返回客户端结果
}

func (m *LoginModule) RspToken(ctx context.Context, req *netpack.G2L_Login) *netpack.L2G_Login {
	token := util.GetUUID() //创建user对应的token
	xlog.Debug("创建user[%s]对应的token[%s]", req.User, token)
	m.mux.Lock()
	m.user2Token[req.User] = UserToken{user: req.User, token: token, time: time.Now()} //将user、token保存
	m.mux.Unlock()
	return &netpack.L2G_Login{Token: token}
}
