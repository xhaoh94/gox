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
	//LoginModule 登录模块
	LoginModule struct {
		gox.Module
		mux              sync.RWMutex
		user2Token       map[string]UserToken
		sessionId2unitId map[uint32]uint32
	}
	UserToken struct {
		user  string
		token string
		time  time.Time
	}
)

//OnInit 初始化
func (m *LoginModule) OnInit() {
	m.user2Token = make(map[string]UserToken)
	m.sessionId2unitId = make(map[uint32]uint32)
	m.RegisterRPC(m.RspToken)
	m.Register(netpack.CMD_C2L_Login, m.RspLogin)
	m.Register(netpack.CMD_C2L_Enter, m.RspEnter)

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
	m.sessionId2unitId[session.ID()] = util.StringToHash(req.User)   //将连接id和玩家绑定
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

func (m *LoginModule) RspEnter(ctx context.Context, session types.ISession, req *netpack.C2L_Enter) {

	unitId := m.sessionId2unitId[session.ID()] //取出玩家id
	sId := req.SceneId
	backRsp := &netpack.S2L_Enter{}
	b := m.GetActorCtrl().Call(uint32(sId), &netpack.L2S_Enter{UnitId: uint(unitId)}, backRsp).Await() //Actor玩家进入场景

	enterRsp := &netpack.L2C_Enter{}
	if b { //玩家进入场景成功
		rsp := &netpack.S2L_SayHello{}
		b = m.GetActorCtrl().Call(unitId, &netpack.L2S_SayHello{Txt: "你好啊，我是机器人"}, rsp).Await() //Actor 玩家发言
		if b {
			xlog.Debug("发言返回:%s", rsp.BackTxt)
			enterRsp.Code = 0
		} else {
			enterRsp.Code = 2
		}
	} else {
		enterRsp.Code = 1
	}
	session.Send(netpack.CMD_L2C_Enter, enterRsp)
}
