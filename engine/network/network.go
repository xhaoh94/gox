package network

import (
	"context"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	NetWork struct {
		context   context.Context
		contextFn context.CancelFunc

		engine           types.IEngine
		outside          types.IService
		interior         types.IService
		rpc              types.IRPC
		serviceDiscovery types.IServiceDiscovery
		actorDiscovery   types.IActorDiscovery
		cmdType          map[uint32]reflect.Type
		cmdLock          sync.RWMutex
	}
)

func New(engine types.IEngine, ctx context.Context) *NetWork {
	nw := new(NetWork)
	nw.engine = engine
	nw.cmdType = make(map[uint32]reflect.Type)
	nw.context, nw.contextFn = context.WithCancel(ctx)
	nw.rpc = rpc.New(engine)
	nw.actorDiscovery = newActorDiscovery(engine, nw.context)
	nw.serviceDiscovery = newServiceDiscovery(engine, nw.context)
	return nw
}

func (nw *NetWork) Engine() types.IEngine {
	return nw.engine
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
func (nw *NetWork) Rpc() types.IRPC {
	return nw.rpc
}
func (nw *NetWork) ServiceDiscovery() types.IServiceDiscovery {
	return nw.serviceDiscovery
}
func (nw *NetWork) ActorDiscovery() types.IActorDiscovery {
	return nw.actorDiscovery
}
func (nw *NetWork) OutsideAddr() string {
	if nw.outside != nil {
		return nw.outside.GetAddr()
	}
	return ""
}
func (nw *NetWork) InteriorAddr() string {
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
	nw.rpc.(*rpc.RPC).Start()
	nw.serviceDiscovery.(*ServiceDiscovery).Start()
	nw.actorDiscovery.(*ActorDiscovery).Start()
}
func (nw *NetWork) Destroy() {
	nw.contextFn()
	if nw.outside != nil {
		nw.outside.Stop()
	}
	nw.interior.Stop()
	nw.rpc.(*rpc.RPC).Stop()
	nw.serviceDiscovery.(*ServiceDiscovery).Stop()
	nw.actorDiscovery.(*ActorDiscovery).Stop()
}
func (nw *NetWork) Serve() {
	nw.rpc.(*rpc.RPC).Serve()
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

// SetGrpcAddr 设置grpc服务
func (nw *NetWork) SetGrpcAddr(addr string) {
	nw.rpc.(*rpc.RPC).SetAddr(addr)
}
