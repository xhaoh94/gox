package mods

import (
	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/examples/uxgame/game"
	"github.com/xhaoh94/gox/examples/uxgame/mods/gate"
	"github.com/xhaoh94/gox/examples/uxgame/mods/login"
	"github.com/xhaoh94/gox/examples/uxgame/mods/scene"
)

type (
	//MainModule 主模块
	MainModule struct {
		gox.Module
	}
)

func (m *MainModule) OnInit() {
	switch gox.Config.AppType {
	case game.Gate:
		m.Put(&gate.GateModule{})
	case game.Login:
		m.Put(&login.LoginModule{})
	case game.Scene:
		m.Put(&scene.SceneModule{})
	default:
		m.Put(&gate.GateModule{})
		m.Put(&login.LoginModule{})
		m.Put(&scene.SceneModule{})
	}
}
