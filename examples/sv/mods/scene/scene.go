package scene

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/actor"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/examples/netpack"
	"github.com/xhaoh94/gox/examples/sv/game"
)

type (
	//SceneModule 场景
	SceneModule struct {
		gox.Module
		scenes map[uint]*Scene
	}

	Scene struct {
		actor.Actor
		Id    uint
		units map[uint]*Unit
	}
)

func newScene(id uint) *Scene {
	scene := &Scene{Id: id, units: make(map[uint]*Unit)}
	scene.OnInit()
	game.Engine.GetNetWork().GetActorCtrl().Add(scene) //把场景添加进Actor
	return scene
}

//OnStart 初始化
func (m *SceneModule) OnStart() {
	m.scenes = make(map[uint]*Scene)
	scene1 := newScene(1) //创建场景1
	m.scenes[scene1.Id] = scene1
	// scene2 := newScene(2) //创建场景2
	// m.scenes[scene2.Id] = scene2
}

func (s *Scene) ActorID() uint32 {
	return uint32(s.Id)
}

func (s *Scene) OnInit() {
	s.AddActorFn(s.OnUnitEnter) //添加到Actor回调
}

func (s *Scene) OnUnitEnter(ctx context.Context, req *netpack.L2S_Enter) *netpack.S2L_Enter {
	xlog.Debug("有玩家进入unitId:%d", req.UnitId)
	unit := newUnit(req.UnitId) //创建玩家
	unit.OnInit()
	game.Engine.GetNetWork().GetActorCtrl().Add(unit) //添加到Actor
	return &netpack.S2L_Enter{Code: 0}
}
