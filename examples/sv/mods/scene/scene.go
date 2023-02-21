package scene

import (
	"context"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
)

type (
	//SceneModule 场景
	SceneModule struct {
		gox.Module
		scenes map[uint]*Scene
	}

	Scene struct {
		location.Entity
		Id    uint
		units map[uint]*Unit
		mux   sync.Mutex
	}
)

// OnInit 初始化
func (m *SceneModule) OnInit() {

}
func (m *SceneModule) OnStart() {
	m.scenes = make(map[uint]*Scene)
	scene1 := newScene(1) //创建场景1
	m.scenes[scene1.Id] = scene1
	scene2 := newScene(2) //创建场景2
	m.scenes[scene2.Id] = scene2
}
func newScene(id uint) *Scene {
	scene := &Scene{Id: id, units: make(map[uint]*Unit)}
	gox.Location.Add(scene) //把场景添加进Actor
	return scene
}

func (s *Scene) LocationID() uint32 {
	return uint32(s.Id)
}

func (s *Scene) OnInit() {
	protoreg.AddLocationRpc(s, s.OnUnitEnter)
}

func (s *Scene) OnUnitEnter(ctx context.Context, session types.ISession, req *netpack.L2S_Enter) (*netpack.S2L_Enter, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.units[req.UnitId]; ok {
		return &netpack.S2L_Enter{Code: 0}, nil
	}
	xlog.Debug("有玩家进入unitId:%d", req.UnitId)
	unit := newUnit(req.UnitId) //创建玩家
	s.units[unit.Id] = unit
	return &netpack.S2L_Enter{Code: 0}, nil
}
