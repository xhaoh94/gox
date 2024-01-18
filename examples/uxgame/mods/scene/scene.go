package scene

import (
	"context"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/aoi"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
	"github.com/xhaoh94/gox/examples/uxgame/game"
)

type (
	//MapModule 地图模块
	SceneModule struct {
		gox.Module
		scenes map[uint]*Scene
	}

	Scene struct {
		aoiMgr types.IAOIManager[uint32]
		location.Location
		Id     uint
		units  map[uint32]*Unit
		mux    sync.RWMutex
		Width  int
		Height int
	}
)

// OnInit 初始化
func (m *SceneModule) OnInit() {

}

func (m *SceneModule) OnStart() {
	m.scenes = make(map[uint]*Scene)
	scene := newScene(1) //创建场景1
	m.scenes[scene.Id] = scene
}

func newScene(id uint) *Scene {
	scene := &Scene{
		Id:     id,
		units:  make(map[uint32]*Unit),
		aoiMgr: aoi.NewAOIGridManager[uint32](-12.5, 12.5, -12.5, 12.5, 1, 1, 5),
		Width:  25,
		Height: 25,
	}
	gox.Location.Register(scene) //把场景添加进Actor
	return scene
}

func (s *Scene) LocationID() uint32 {
	return uint32(s.Id)
}

func (s *Scene) OnInit() {
	protoreg.AddLocationRpc(s, s.OnEnterScene)
}

func (s *Scene) OnEnterScene(ctx context.Context, session types.ISession, req *pb.C2S_EnterScene) (*pb.S2C_EnterScene, error) {

	rId := strhelper.StringToHash(req.Account)
	s.mux.RLock()
	_, ok := s.units[rId]
	s.mux.RUnlock()

	if ok {
		return &pb.S2C_EnterScene{Error: pb.ErrCode_UnKnown}, nil
	}

	logger.Debug().Msgf("玩家进入account:%s,roleID:%d", req.Account, rId)
	entity := &pb.Entity{
		RoleId:   rId,
		RoleMask: req.RoleMask,
		Position: &pb.Vector3{X: 0, Y: 0, Z: 0},
	}
	unit := newUnit(entity, session) //创建玩家

	s.mux.Lock()
	s.units[rId] = unit
	unit.EnterScene(s)
	s.mux.Unlock()

	aoiResult := unit.GetAOIResult()
	if aoiResult != nil {
		ids := aoiResult.IDList()
		defer s.interiorUnitIntoView(ids, unit)
		return &pb.S2C_EnterScene{Self: entity, Others: s.GetEntitys(ids)}, nil
	}
	return &pb.S2C_EnterScene{Error: pb.ErrCode_UnKnown}, nil
}

func (s *Scene) OnLeaveScene(unit *Unit) {
	s.mux.Lock()
	delete(s.units, unit.UnitID)
	s.mux.Unlock()

	aoiResult := unit.GetAOIResult()
	if aoiResult != nil {
		sessions := s.GetSessions(aoiResult.IDList())
		if len(sessions) == 0 {
			return
		}
		req := &pb.Bcst_UnitOutofView{
			RoleId: unit.UnitID,
		}
		s.Bcst(sessions, pb.CMD_Bcst_UnitOutofView, req)
	}
}

func (s *Scene) interiorUnitIntoView(ids []uint32, unit *Unit) {
	if len(ids) == 0 {
		return
	}
	sessions := s.GetSessions(ids)
	if len(sessions) == 0 {
		return
	}
	req := &pb.Bcst_UnitIntoView{
		Role: &pb.Entity{
			RoleId:   unit.UnitID,
			RoleMask: unit.UnitMask,
			Position: &pb.Vector3{
				X: unit.Position.X,
				Y: unit.Position.Y,
				Z: unit.Position.Z,
			},
		},
	}
	s.Bcst(sessions, pb.CMD_Bcst_UnitIntoView, req)
}

func (s *Scene) interiorUnitOutofView(ids []uint32, unit *Unit) {
	if len(ids) == 0 {
		return
	}
	sessions := s.GetSessions(ids)
	if len(sessions) == 0 {
		return
	}
	req := &pb.Bcst_UnitOutofView{
		RoleId: unit.UnitID,
	}
	s.Bcst(sessions, pb.CMD_Bcst_UnitOutofView, req)
}

func (s *Scene) interiorUnitMove(ids []uint32, unit *Unit, points []*pb.Vector3, index int) {
	if len(ids) == 0 {
		return
	}
	sessions := s.GetSessions(ids)
	if len(sessions) == 0 {
		return
	}
	req := &pb.Bcst_UnitMove{
		RoleId:     unit.UnitID,
		PointIndex: int32(index),
		Points:     points,
	}
	s.Bcst(sessions, pb.CMD_Bcst_UnitMove, req)
}
func (s *Scene) interiorUnitPosition(ids []uint32, unit *Unit) {
	if len(ids) == 0 {
		return
	}
	sessions := s.GetSessions(ids)
	if len(sessions) == 0 {
		return
	}
	req := &pb.Bcst_UnitUpdatePosition{
		RoleId: unit.UnitID,
		Point: &pb.Vector3{
			X: unit.Position.X,
			Y: unit.Position.Y,
			Z: unit.Position.Z,
		},
	}
	s.Bcst(sessions, pb.CMD_Bcst_UnitUpdatePosition, req)
}

func (s *Scene) Bcst(sessions map[uint32][]uint32, cmd uint32, require any) {
	if len(sessions) == 0 {
		return
	}
	for sid, roles := range sessions {
		session := gox.NetWork.GetSessionById(sid)
		if session == nil {
			continue
		}
		var datas []byte
		if require != nil {
			var err error
			if datas, err = session.Codec(cmd).Marshal(require); err != nil {
				logger.Err(err).Msg("广播转发失败")
				continue
			}
		}
		relay := &game.Interior_Relay{
			Roles:   roles,
			CMD:     cmd,
			Require: datas,
		}
		session.Send(game.InteriorRelay, relay)
	}
}

func (s *Scene) GetEntitys(ids []uint32) []*pb.Entity {
	defer s.mux.RUnlock()
	s.mux.RLock()

	entitys := make([]*pb.Entity, 0)
	for _, id := range ids {
		unit, ok := s.units[id]
		if !ok {
			continue
		}
		entitys = append(entitys, &pb.Entity{
			RoleId:   unit.UnitID,
			RoleMask: unit.UnitMask,
			Position: &pb.Vector3{
				X: unit.Position.X,
				Y: unit.Position.Y,
				Z: unit.Position.Z,
			},
		})
	}
	return entitys
}
func (s *Scene) GetSessions(ids []uint32) map[uint32][]uint32 {
	defer s.mux.RUnlock()
	s.mux.RLock()

	sessionRoles := make(map[uint32][]uint32)
	for _, id := range ids {
		unit, ok := s.units[id]
		if !ok {
			continue
		}

		if roles, ok := sessionRoles[unit.GateSession]; ok {
			roles = append(roles, unit.UnitID)
			sessionRoles[unit.GateSession] = roles
		} else {
			roles = []uint32{unit.UnitID}
			sessionRoles[unit.GateSession] = roles
		}
	}
	return sessionRoles
}
