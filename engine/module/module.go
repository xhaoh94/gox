package module

import (
	"github.com/xhaoh94/gox/engine/event"
	"github.com/xhaoh94/gox/engine/xlog"
)

var (
	mainModule IModule
	isRun      bool
)

//SetModule 设置主模块
func SetModule(m IModule) {
	mainModule = m
}

//Start 模块初始化
func Start() {
	if isRun {
		return
	}
	if mainModule == nil {
		xlog.Error("mainModule is nil!")
		return
	}
	isRun = true
	mainModule.Init(mainModule)
	event.Call("_init_module_ok_")
}

//Stop 模块销毁
func Stop() {
	if !isRun {
		return
	}
	if mainModule != nil {
		mainModule.Destroy(mainModule)
	}
	isRun = false
}
