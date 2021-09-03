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
	switch m.GetEngine().ServiceType() {
	case game.Gate:
		m.Put(&gate.GateModule{})
		break
	case game.Login:
		m.Put(&login.LoginModule{})
		break
	case game.Scene:
		m.Put(&scene.SceneModule{})
		break
	default:
		m.Put(&gate.GateModule{})
		m.Put(&login.LoginModule{})
		m.Put(&scene.SceneModule{})
		break
	}
}

//OnStart 初始化
func (m *MainModule) OnStart() {

}
