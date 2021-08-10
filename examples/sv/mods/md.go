package mods

import (
	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/examples/sv/mods/gate"
	"github.com/xhaoh94/gox/examples/sv/mods/login"
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
		gox.Module
	}
)

//OnInit 初始化
func (mm *MainModule) OnInit() {
	switch mm.GetEngine().ServiceType() {
	case Gate:
		mm.Put(&gate.GateModule{})
		break
	case Login:
		mm.Put(&login.LoginModule{})
		break
	}
}
