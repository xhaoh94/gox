package network

import (
	"context"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/discovery"

	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/helper/commonhelper"
	"github.com/xhaoh94/gox/types"
)

type (
	NetWork struct {
		engine   types.IEngine
		outside  types.IService
		interior types.IService

		context   context.Context
		contextFn context.CancelFunc

		actorDiscovery   *discovery.ActorDiscovery
		serviceDiscovery *discovery.ServiceDiscovery
		cmdType          map[uint32]reflect.Type
		cmdLock          sync.RWMutex
	}
)

func New(engine types.IEngine, ctx context.Context) *NetWork {
	nw := new(NetWork)
	nw.engine = engine
	nw.actorDiscovery = discovery.NewActorDiscovery(engine)
	nw.serviceDiscovery = discovery.NewServiceDiscovery(engine)
	nw.cmdType = make(map[uint32]reflect.Type)
	nw.context, nw.contextFn = context.WithCancel(ctx)
	return nw
}

func (nw *NetWork) Engine() types.IEngine {
	return nw.engine
}
func (nw *NetWork) ServiceDiscovery() types.IServiceDiscovery {
	return nw.serviceDiscovery
}
func (nw *NetWork) ActorDiscovery() types.IActorDiscovery {
	return nw.actorDiscovery
}

// GetSession 通过id获取Session
func (nw *NetWork) GetSessionById(sid uint32) types.ISession {
	session := nw.interior.GetSessionById(sid)
	if session == nil && nw.outside != nil {
		session = nw.outside.GetSessionById(sid)
	}
	return session
}

// GetSessionByAddr 通过地址获取Session
func (nw *NetWork) GetSessionByAddr(addr string) types.ISession {
	return nw.interior.GetSessionByAddr(addr)
}

func (nw *NetWork) GetOutsideAddr() string {
	if nw.outside != nil {
		return nw.outside.GetAddr()
	}
	return ""
}
func (nw *NetWork) GetInteriorAddr() string {
	if nw.interior != nil {
		return nw.interior.GetAddr()
	}
	return ""
}

// RegisterRType 注册协议消息体类型
func (nw *NetWork) RegisterRType(cmd uint32, protoType reflect.Type) {
	defer nw.cmdLock.Unlock()
	nw.cmdLock.Lock()
	if _, ok := nw.cmdType[cmd]; ok {
		xlog.Error("重复注册协议 cmd[%s]", cmd)
		return
	}
	nw.cmdType[cmd] = protoType
}

// RegisterRType 注册协议消息体类型
func (nw *NetWork) UnRegisterRType(cmd uint32) {
	defer nw.cmdLock.Unlock()
	nw.cmdLock.Lock()
	if _, ok := nw.cmdType[cmd]; ok {
		delete(nw.cmdType, cmd)
	}
}

// GetRegProtoMsg 获取协议消息体
func (nw *NetWork) GetRegProtoMsg(cmd uint32) interface{} {
	nw.cmdLock.RLock()
	rType, ok := nw.cmdType[cmd]
	nw.cmdLock.RUnlock()
	if !ok {
		return nil
	}
	return commonhelper.RTypeToInterface(rType)
}

func (nw *NetWork) Init() {

	if nw.interior == nil {
		xlog.Fatal("没有初始化内部网络通信")
		return
	}
	nw.interior.Start()
	if nw.outside != nil {
		nw.outside.Start()
	}
	nw.serviceDiscovery.Start(nw.context)
	nw.actorDiscovery.Start(nw.context)
}
func (nw *NetWork) Destroy() {
	nw.contextFn()
	nw.actorDiscovery.Stop()
	nw.serviceDiscovery.Stop()
	if nw.outside != nil {
		nw.outside.Stop()
	}
	nw.interior.Stop()
}

// SetOutsideService 设置外部服务类型
func (nw *NetWork) SetOutsideService(ser types.IService, addr string, codec types.ICodec) {
	if addr == "" {
		return
	}
	nw.outside = ser
	nw.outside.Init(addr, codec, nw.engine, nw.context)
}

// SetInteriorService 设置内部服务类型
func (nw *NetWork) SetInteriorService(ser types.IService, addr string, codec types.ICodec) {
	if addr == "" {
		return
	}
	nw.interior = ser
	nw.interior.Init(addr, codec, nw.engine, nw.context)
}
