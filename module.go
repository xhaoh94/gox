package gox

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
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
func (m *Module) Start(self types.IModule, engine types.IEngine) {
	m.engine = engine
	self.OnInit()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Start(v, engine)
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

//OnDestroy 进行销毁
func (m *Module) OnDestroy() {

}
func (m *Module) GetEngine() types.IEngine {
	return m.engine
}

func (m *Module) RegisterActor(actorID uint32, sessionID string) {
	m.engine.GetNetWork().GetActor().Register(actorID, sessionID)
}
func (m *Module) Actor(actorID uint32, cmd uint32, msg interface{}) {
	m.engine.GetNetWork().GetActor().Send(actorID, cmd, msg)
}
func (m *Module) GetSessionById(sid string) types.ISession {
	return m.engine.GetNetWork().GetSessionById(sid)
}
func (m *Module) GetSessionByAddr(addr string) types.ISession {
	return m.engine.GetNetWork().GetSessionByAddr(addr)
}
func (m *Module) GetGrpcConnByAddr(addr string) *grpc.ClientConn {
	return m.engine.GetNetWork().GetRPC().GetConnByAddr(addr)
}
func (m *Module) GetGrpcServer() *grpc.Server {
	return m.engine.GetNetWork().GetRPC().GetServer()
}

//Register 注册协议对应消息体和回调函数
func (m *Module) Register(cmd uint32, fn interface{}) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	switch tFun.NumIn() {
	case 3:
		in := tFun.In(2)
		if in.Kind() != reflect.Ptr {
			xlog.Error("注册协议回调函数参数结构体需要是指针类型 cmd[%d]", cmd)
			return
		}
		m.engine.GetNetWork().RegisterRType(cmd, in)
		break
	case 2:
		break
	default:
		xlog.Error("注册协议回调函数参数有误")
		return
	}
	m.engine.GetEvent().Bind(cmd, fn)
}

//RegisterRPC 注册rpc
func (m *Module) RegisterRPC(args ...interface{}) {
	l := len(args)
	var msgid uint32
	var fn interface{}
	switch l {
	case 1:
		fn = args[0]
		break
	case 2:
		msgid = uint32(args[0].(int))
		fn = args[1]
		break
	default:
		xlog.Error("注册协议回调函数参数有误")
		return
	}
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	if out.Kind() != reflect.Ptr {
		xlog.Error("注册prc回调函数参数结构体需要是指针类型")
		return
	}
	if tFun.NumIn() == 2 {
		in := tFun.In(1)
		if in.Kind() != reflect.Ptr {
			xlog.Error("注册prc回调函数参数结构体需要是指针类型")
			return
		}
		if msgid == 0 {
			key := in.Elem().Name() + out.Elem().Name()
			msgid = util.StringToHash(key)
		}
		m.engine.GetNetWork().RegisterRType(msgid, in)
		m.engine.GetEvent().Bind(msgid, fn)
		return
	}
	if msgid == 0 {
		key := out.Elem().Name()
		msgid = util.StringToHash(key)
	}
	m.engine.GetEvent().Bind(msgid, fn)
}
