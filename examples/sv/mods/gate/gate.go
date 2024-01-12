package gate

import (
	"context"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
)

type (
	//GateModule 网关
	GateModule struct {
		gox.Module
		muxToken    sync.RWMutex
		userToken   map[string]UserToken
		muxSession  sync.RWMutex
		roleSession map[uint32]uint32
		sessionRole map[uint32]uint32
	}
	UserToken struct {
		user  string
		token string
		time  time.Time
	}
)

// OnInit 初始化
func (m *GateModule) OnInit() {
	m.userToken = make(map[string]UserToken)
	m.roleSession = make(map[uint32]uint32)
	m.sessionRole = make(map[uint32]uint32)
	pb.RegisterILoginGameServer(gox.NetWork.Rpc().GRpcServer(), m)
	protoreg.RegisterRpcCmd(pb.CMD_C2S_EnterMap, m.EnterGame)
	protoreg.Register(pb.CMD_C2S_Move, m.Move)

	protoreg.Register(pb.CMD_Interior_EnterVision, m.InteriorEnterVision)
	protoreg.Register(pb.CMD_Interior_LeaveVision, m.InteriorLeaveVision)
	protoreg.Register(pb.CMD_Interior_Move, m.InteriorMove)

	gox.NetWork.Outside().LinstenByDelSession(m.OnSessionStop)
}
func (m *GateModule) OnSessionStop(sid uint32) {
	m.muxSession.Lock()
	if rid, ok := m.sessionRole[sid]; ok {
		delete(m.sessionRole, sid)
		delete(m.roleSession, rid)
		gox.Location.Send(rid, &pb.C2S_LeaveMap{RoleId: rid})
	}
	m.muxSession.Unlock()
}

func (m *GateModule) LoginGame(ctx context.Context, req *pb.C2S_LoginGame) (*pb.S2C_LoginGame, error) {
	token := commonhelper.NewUUID() //创建user对应的token
	logger.Debug().Msgf("创建账号[%s]对应的token[%s]", req.Account, token)
	m.muxToken.Lock()
	m.userToken[req.Account] = UserToken{user: req.Account, token: token, time: time.Now()} //将user、token保存
	m.muxToken.Unlock()

	return &pb.S2C_LoginGame{Token: token, Addr: gox.Config.OutsideAddr}, nil
}

func (m *GateModule) EnterGame(ctx context.Context, session types.ISession, req *pb.C2S_EnterMap) (*pb.S2C_EnterMap, error) {
	m.muxToken.RLock()
	ut, ok := m.userToken[req.Account]
	m.muxToken.RUnlock()
	resp := &pb.S2C_EnterMap{}
	if !ok {
		resp.Error = pb.ErrCode_UnKnown //没有找到对应的token
		return resp, nil
	}
	if ut.token != req.Token {
		resp.Error = pb.ErrCode_UnKnown //token 不一致
		return resp, nil
	}
	defer func() {
		m.muxToken.Lock()
		delete(m.userToken, req.Account)
		m.muxToken.Unlock()
	}()
	t := time.Now().Sub(ut.time)
	if t.Seconds() > 5 {
		resp.Error = pb.ErrCode_UnKnown ///token已过期
		return resp, nil
	}

	err := gox.Location.Call(uint32(req.Mapid), req, resp)

	if err != nil {
		logger.Info().Err(err)
		resp.Error = pb.ErrCode_UnKnown
		logger.Err(err)
	} else {
		logger.Info().Msgf("返回进入地图:%d", resp.Self.RoleId)
		m.muxSession.Lock()
		m.roleSession[resp.Self.RoleId] = session.ID()
		m.sessionRole[session.ID()] = resp.Self.RoleId
		m.muxSession.Unlock()
	}
	return resp, nil
}

func (m *GateModule) Move(ctx context.Context, session types.ISession, req *pb.C2S_Move) {
	m.muxSession.RLock()
	defer m.muxSession.RUnlock()
	if rid, ok := m.sessionRole[session.ID()]; ok {
		gox.Location.Send(rid, req)
	}
}

func (m *GateModule) InteriorEnterVision(ctx context.Context, session types.ISession, req *pb.Interior_EnterVision) {

	defer m.muxSession.RUnlock()
	m.muxSession.RLock()
	logger.Debug().Msgf("广播进入视野:%d", req.Role.RoleId)
	bcstResp := &pb.Bcst_EnterMap{
		Role: req.Role,
	}
	for _, roleId := range req.Roles {
		if sid, ok := m.roleSession[roleId]; ok {
			_session := gox.NetWork.GetSessionById(sid)
			_session.Send(pb.CMD_Bcst_EnterMap, bcstResp)
		}
	}
}
func (m *GateModule) InteriorLeaveVision(ctx context.Context, session types.ISession, req *pb.Interior_LeaveVision) {

	defer m.muxSession.RUnlock()
	m.muxSession.RLock()
	logger.Debug().Msgf("广播离开视野:%d", req.RoleId)
	bcstResp := &pb.Bcst_LeaveMap{
		RoleId: req.RoleId,
	}
	for _, roleId := range req.Roles {
		if sid, ok := m.roleSession[roleId]; ok {
			_session := gox.NetWork.GetSessionById(sid)
			_session.Send(pb.CMD_Bcst_LeaveMap, bcstResp)
		}
	}
}
func (m *GateModule) InteriorMove(ctx context.Context, session types.ISession, req *pb.Interior_Move) {
	defer m.muxSession.RUnlock()
	m.muxSession.RLock()
	logger.Debug().Msgf("广播移动:%d", req.RoleId)
	bcstResp := &pb.Bcst_Move{
		RoleId: req.RoleId,
		Points: req.Points,
	}
	for _, roleId := range req.Roles {
		if sid, ok := m.roleSession[roleId]; ok {
			_session := gox.NetWork.GetSessionById(sid)
			_session.Send(pb.CMD_Bcst_Move, bcstResp)
		}
	}
}
