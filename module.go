package gox

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
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

// Init 初始化模块
func (m *Module) Init(self types.IModule, engine types.IEngine, fn func()) {
	m.engine = engine
	self.OnInit()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Init(v, engine, nil)
		}
	}
	if fn != nil {
		fn()
		if m.childModules != nil {
			for i := range m.childModules {
				v := m.childModules[i]
				v.OnStart()
			}
		}
	}
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

// Destroy 销毁模块
func (m *Module) Destroy(self types.IModule) {
	for i := range m.childModules {
		v := m.childModules[i]
		v.Destroy(v)
	}
	self.OnDestroy()
}

// OnStop 模块关闭
func (mm *Module) OnDestroy() {

}

func (m *Module) GetEngine() types.IEngine {
	return m.engine
}
func (m *Module) GetActorCtrl() types.IActorDiscovery {
	return m.engine.NetWork().ActorDiscovery()
}

func (m *Module) GetSessionById(sid uint32) types.ISession {
	return m.engine.NetWork().GetSessionById(sid)
}
func (m *Module) GetSessionByAddr(addr string) types.ISession {
	return m.engine.NetWork().GetSessionByAddr(addr)
}
func (m *Module) GetGrpcConnByAddr(addr string) *grpc.ClientConn {
	return m.engine.NetWork().Rpc().GetConnByAddr(addr)
}
func (m *Module) GetGrpcServer() *grpc.Server {
	return m.engine.NetWork().Rpc().GetServer()
}
func (m *Module) GetServiceConfListByType(sType string) []types.IServiceEntity {
	return m.engine.NetWork().ServiceDiscovery().GetServiceConfListByType(sType)
}
func (m *Module) GetServiceConfByID(id uint) types.IServiceEntity {
	return m.engine.NetWork().ServiceDiscovery().GetServiceConfByID(id)
}

// Register 注册协议对应消息体和回调函数
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
		m.engine.NetWork().RegisterRType(cmd, in)
		break
	default:
		xlog.Error("协议回调函数参数有误")
		return
	}
	m.engine.Event().Bind(cmd, fn)
}

// RegisterRPC 注册RPC
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
			cmd = strhelper.StringToHash(key)
		}
		m.engine.NetWork().RegisterRType(cmd, in)
		break
	default:
		xlog.Error("RPC回调函数参数有误")
		return
	}

	if cmd == 0 {
		cmd = strhelper.StringToHash(key)
	}
	m.engine.Event().Bind(cmd, fn)
}
