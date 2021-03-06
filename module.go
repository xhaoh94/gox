package gox

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
	"github.com/xhaoh94/gox/xdef"
	"google.golang.org/grpc"
)

type (
	//Module 模块
	Module struct {
		engine       types.IEngine
		childModules []types.IModule
		lock         sync.Mutex
	}
)

//Init 初始化模块
func (m *Module) Init(self types.IModule, engine types.IEngine) {
	m.engine = engine
	m.engine.GetEvent().On(xdef.START_ENGINE_OK, self.OnStart)
	self.OnInit()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Init(v, engine)
		}
	}
}

//Put 添加模块
func (m *Module) Put(mod types.IModule) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if m.childModules == nil {
		m.childModules = make([]types.IModule, 0)
	}
	m.childModules = append(m.childModules, mod)
}

//Destroy 销毁模块
func (m *Module) Destroy(self types.IModule) {
	for i := range m.childModules {
		v := m.childModules[i]
		v.Destroy(v)
	}
	self.OnDestroy()
}

//OnStop 模块关闭
func (mm *Module) OnDestroy() {

}

func (m *Module) GetEngine() types.IEngine {
	return m.engine
}
func (m *Module) GetActorCtrl() types.IActorCtrl {
	return m.engine.GetNetWork().GetActorCtrl()
}

func (m *Module) GetSessionById(sid uint32) types.ISession {
	return m.engine.GetNetWork().GetSessionById(sid)
}
func (m *Module) GetSessionByAddr(addr string) types.ISession {
	return m.engine.GetNetWork().GetSessionByAddr(addr)
}
func (m *Module) GetGrpcConnByAddr(addr string) *grpc.ClientConn {
	return m.engine.GetRPC().GetConnByAddr(addr)
}
func (m *Module) GetGrpcServer() *grpc.Server {
	return m.engine.GetRPC().GetServer()
}
func (m *Module) GetServiceConfListByType(sType string) []types.IServiceConfig {
	return m.engine.GetNetWork().GetServiceCtrl().GetServiceConfListByType(sType)
}
func (m *Module) GetServiceConfByID(id uint) types.IServiceConfig {
	return m.engine.GetNetWork().GetServiceCtrl().GetServiceConfByID(id)
}

//Register 注册协议对应消息体和回调函数
func (m *Module) Register(cmd uint32, fn interface{}) {

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
		m.engine.GetNetWork().RegisterRType(cmd, in)
		break
	default:
		xlog.Error("协议回调函数参数有误")
		return
	}
	m.engine.GetEvent().Bind(cmd, fn)
}

//RegisterRPC 注册RPC
func (m *Module) RegisterRPC(args ...interface{}) {
	l := len(args)
	var cmd uint32
	var fn interface{}
	switch l {
	case 1:
		fn = args[0]
		break
	case 2:
		cmd = uint32(args[0].(int))
		fn = args[1]
		break
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
			cmd = util.StringToHash(key)
		}
		m.engine.GetNetWork().RegisterRType(cmd, in)
		break
	default:
		xlog.Error("RPC回调函数参数有误")
		return
	}

	if cmd == 0 {
		cmd = util.StringToHash(key)
	}
	m.engine.GetEvent().Bind(cmd, fn)
}
