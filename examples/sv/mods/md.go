package mods

import (
	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/examples/sv/game"
	"github.com/xhaoh94/gox/examples/sv/mods/gate"
	"github.com/xhaoh94/gox/examples/sv/mods/login"
	"github.com/xhaoh94/gox/examples/sv/mods/scene"
)

type (
	//MainModule 主模块
	MainModule struct {
		gox.Module
	}
)

func (m *MainModule) OnInit() {
	switch m.Engine.AppConf().EType {
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

// OnStart 初始化
func (m *MainModule) OnStart() {

}
