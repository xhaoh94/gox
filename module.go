package gox

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	//Module 模块
	Module struct {
		childModules []types.IModule
		lock         sync.Mutex
	}
)

// Init 初始化模块
func (m *Module) Init(self types.IModule) {
	self.OnInit()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Init(v)
		}
	}
}

// OnInit 初始模块
func (mm *Module) OnInit() {

}

func (m *Module) Start(self types.IModule) {
	self.OnStart()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Start(self)
		}
	}
}

// OnStart 模块启动
func (mm *Module) OnStart() {

}

// Destroy 销毁模块
func (m *Module) Destroy(self types.IModule) {
	for i := range m.childModules {
		v := m.childModules[i]
		v.Destroy(v)
	}
	self.OnDestroy()
}

// OnDestroy 模块销毁
func (mm *Module) OnDestroy() {

}

// Put 添加模块
func (m *Module) Put(mod types.IModule) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if m.childModules == nil {
		m.childModules = make([]types.IModule, 0)
	}
	m.childModules = append(m.childModules, mod)
}

// Register 注册协议对应消息体和回调函数
func (m *Module) Register(cmd uint32, fn any) {

	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	if tFun.Kind() != reflect.Func {
		xlog.Error("协议回调函数不是方法 cmd:[%d]", cmd)
		return
	}
	switch tFun.NumIn() {
	case 2: //ctx,session
		break
	case 3: //ctx,session,req
		in := tFun.In(2)
		if in.Kind() != reflect.Ptr {
			xlog.Error("协议回调函数参数需要是指针类型 cmd[%d]", cmd)
			return
		}
		protoreg.RegisterRType(cmd, in)
	default:
		xlog.Error("协议回调函数参数有误")
		return
	}
	Event.Bind(cmd, fn)
}

// RegisterRpc 注册RPC
func (m *Module) RegisterRpc(args ...interface{}) {
	l := len(args)
	var cmd uint32
	var fn interface{}
	switch l {
	case 1:
		fn = args[0]
	case 2:
		cmd = uint32(args[0].(int))
		fn = args[1]
	default:
		xlog.Error("RPC回调函数参数有误")
		return
	}
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	if tFun.Kind() != reflect.Func {
		xlog.Error("RPC回调函数不是方法")
		return
	}
	if tFun.NumOut() != 1 {
		xlog.Error("RPC回调函数参数有误")
		return
	}
	out := tFun.Out(0)
	if out.Kind() != reflect.Ptr {
		xlog.Error("RPC函数参数需要是指针类型")
		return
	}

	key := out.Elem().Name()

	switch tFun.NumIn() {
	case 1: //ctx
		break
	case 2: //ctx,req
		in := tFun.In(1)
		if in.Kind() != reflect.Ptr {
			xlog.Error("RPC函数参数需要是指针类型")
			return
		}
		if cmd == 0 {
			key = in.Elem().Name() + key
			cmd = strhelper.StringToHash(key)
		}
		protoreg.RegisterRType(cmd, in)
	default:
		xlog.Error("RPC回调函数参数有误")
		return
	}

	if cmd == 0 {
		cmd = strhelper.StringToHash(key)
	}
	Event.Bind(cmd, fn)
}
