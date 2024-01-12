package scene

import (
	"context"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
)

type (
	//MapModule 地图模块
	SceneModule struct {
		gox.Module
		scenes map[uint]*Scene
	}

	Scene struct {
		location.Location
		Id    uint
		units map[uint32]*Unit
		mux   sync.RWMutex
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
	scene := &Scene{Id: id, units: make(map[uint32]*Unit)}
	gox.Location.Register(scene) //把场景添加进Actor
	return scene
}

func (s *Scene) LocationID() uint32 {
	return uint32(s.Id)
}

func (s *Scene) OnInit() {
	protoreg.AddLocationRpc(s, s.OnEnterMap)
}

func (s *Scene) OnEnterMap(ctx context.Context, session types.ISession, req *pb.C2S_EnterMap) (*pb.S2C_EnterMap, error) {

	s.mux.RLock()
	rId := strhelper.StringToHash(req.Account)
	if _, ok := s.units[rId]; ok {
		s.mux.RUnlock()
		return &pb.S2C_EnterMap{Error: pb.ErrCode_UnKnown}, nil
	}

	logger.Debug().Msgf("玩家进入account:%s,roleID:%d", req.Account, rId)
	entity := &pb.Entity{
		RoleId:   rId,
		RoleMask: req.RoleMask,
		Position: &pb.Vector3{X: 0, Y: 0, Z: 0},
	}
	unit := newUnit(entity, session, s) //创建玩家
	others := make([]*pb.Entity, 0)
	for _, v := range s.units {
		others = append(others, v.Entity)
	}
	s.mux.RUnlock()

	s.mux.Lock()
	s.units[rId] = unit
	s.mux.Unlock()

	defer s.interiorEnterVision(entity)
	return &pb.S2C_EnterMap{Self: entity, Others: others}, nil
}
func (s *Scene) interiorEnterVision(entity *pb.Entity) {
	defer s.mux.RUnlock()
	s.mux.RLock()
	tem := make(map[types.ISession][]uint32)
	for _, v := range s.units {
		if v.Entity == entity {
			continue
		}

		if roles, ok := tem[v.GateSession]; ok {
			roles = append(roles, v.Entity.RoleId)
			tem[v.GateSession] = roles
		} else {
			roles = []uint32{v.Entity.RoleId}
			tem[v.GateSession] = roles
		}
	}

	for session, roles := range tem {
		req := &pb.Interior_EnterVision{
			Roles: roles,
			Role:  entity,
		}
		session.Send(pb.CMD_Interior_EnterVision, req)
	}
}

func (s *Scene) interiorLeaveVision(entity *pb.Entity) {
	s.mux.Lock()
	delete(s.units, entity.RoleId)
	s.mux.Unlock()

	s.mux.RLock()
	defer s.mux.RUnlock()
	tem := make(map[types.ISession][]uint32)
	for _, v := range s.units {
		if v.Entity == entity {
			continue
		}

		if roles, ok := tem[v.GateSession]; ok {
			roles = append(roles, v.Entity.RoleId)
			tem[v.GateSession] = roles
		} else {
			roles = []uint32{v.Entity.RoleId}
			tem[v.GateSession] = roles
		}
	}

	for session, roles := range tem {
		req := &pb.Interior_LeaveVision{
			Roles:  roles,
			RoleId: entity.RoleId,
		}
		session.Send(pb.CMD_Interior_LeaveVision, req)
	}
}

func (s *Scene) interiorMove(entity *pb.Entity, points []*pb.Vector3) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	tem := make(map[types.ISession][]uint32)
	for _, v := range s.units {
		if roles, ok := tem[v.GateSession]; ok {
			roles = append(roles, v.Entity.RoleId)
			tem[v.GateSession] = roles
		} else {
			roles = []uint32{v.Entity.RoleId}
			tem[v.GateSession] = roles
		}
	}

	for session, roles := range tem {
		req := &pb.Interior_Move{
			Roles:  roles,
			RoleId: entity.RoleId,
			Points: points,
		}
		session.Send(pb.CMD_Interior_Move, req)
	}
}
