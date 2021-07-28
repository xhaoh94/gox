package mods

import (
	"github.com/xhaoh94/gox/engine/module"
	"github.com/xhaoh94/gox/game/mods/gate"
	"github.com/xhaoh94/gox/game/mods/login"
)

const (
	//Gate 网关服务
	Gate string = "gate"
	//Login 登录服务
	Login string = "login"
)

type (
	//MainModule 主模块
	MainModule struct {
		module.Module
	}
)

//OnInit 初始化
func (mm *MainModule) OnInit() {
	switch mm.GetEngine().GetServiceType() {
	case Gate:
		mm.Put(&gate.GateModule{})
		break
	case Login:
		mm.Put(&login.LoginModule{})
		break
	}
}
