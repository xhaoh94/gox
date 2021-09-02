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

//OnStart 初始化
func (mm *MainModule) OnStart() {
	switch mm.GetEngine().ServiceType() {
	case game.Gate:
		mm.Put(&gate.GateModule{})
		break
	case game.Login:
		mm.Put(&login.LoginModule{})
		break
	case game.Scene:
		mm.Put(&scene.SceneModule{})
		break
	default:
		mm.Put(&gate.GateModule{})
		mm.Put(&login.LoginModule{})
		break
	}
}
