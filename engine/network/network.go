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
	network := new(NetWork)
	network.engine = engine
	network.cmdType = make(map[uint32]reflect.Type)
	network.context, network.contextFn = context.WithCancel(ctx)
	network.rpc = rpc.New(engine)
	network.actorDiscovery = newActorDiscovery(engine, network.context)
	network.serviceDiscovery = newServiceDiscovery(engine, network.context)
	return network
}

func (network *NetWork) Engine() types.IEngine {
	return network.engine
}

// GetSession 通过id获取Session
func (network *NetWork) GetSessionById(sid uint32) types.ISession {
	session := network.interior.GetSessionById(sid)
	if session == nil && network.outside != nil {
		session = network.outside.GetSessionById(sid)
	}
	return session
}

// GetSessionByAddr 通过地址获取Session
func (network *NetWork) GetSessionByAddr(addr string) types.ISession {
	return network.interior.GetSessionByAddr(addr)
}
func (network *NetWork) Rpc() types.IRPC {
	return network.rpc
}
func (network *NetWork) ServiceDiscovery() types.IServiceDiscovery {
	return network.serviceDiscovery
}
func (network *NetWork) ActorDiscovery() types.IActorDiscovery {
	return network.actorDiscovery
}

// RegisterRType 注册协议消息体类型
func (network *NetWork) RegisterRType(cmd uint32, protoType reflect.Type) {
	defer network.cmdLock.Unlock()
	network.cmdLock.Lock()
	if _, ok := network.cmdType[cmd]; ok {
		xlog.Error("重复注册协议 cmd[%s]", cmd)
		return
	}
	network.cmdType[cmd] = protoType
}

// RegisterRType 注册协议消息体类型
func (network *NetWork) UnRegisterRType(cmd uint32) {
	defer network.cmdLock.Unlock()
	network.cmdLock.Lock()
	delete(network.cmdType, cmd)
}

// GetRegProtoMsg 获取协议消息体
func (network *NetWork) GetRegProtoMsg(cmd uint32) interface{} {
	network.cmdLock.RLock()
	rType, ok := network.cmdType[cmd]
	network.cmdLock.RUnlock()
	if !ok {
		return nil
	}
	return commonhelper.RTypeToInterface(rType)
}

func (network *NetWork) Init() {

	if network.interior == nil {
		xlog.Fatal("没有初始化内部网络通信")
		return
	}
	network.interior.Start()
	if network.outside != nil {
		network.outside.Start()
	}
	network.rpc.(*rpc.RPC).Start()
	network.serviceDiscovery.(*ServiceDiscovery).Start()
	network.actorDiscovery.(*ActorDiscovery).Start()
}
func (network *NetWork) Destroy() {
	network.contextFn()
	if network.outside != nil {
		network.outside.Stop()
	}
	network.interior.Stop()
	network.rpc.(*rpc.RPC).Stop()
	network.serviceDiscovery.(*ServiceDiscovery).Stop()
	network.actorDiscovery.(*ActorDiscovery).Stop()
}
func (network *NetWork) Serve() {
	network.rpc.(*rpc.RPC).Serve()
}

// SetOutsideService 设置外部服务类型
func (network *NetWork) SetOutsideService(ser types.IService, codec types.ICodec) {
	addr := network.engine.AppConf().OutsideAddr
	if addr == "" {
		return
	}
	ser.Init(addr, codec, network.engine, network.context)
	network.outside = ser
}

// SetInteriorService 设置内部服务类型
func (network *NetWork) SetInteriorService(ser types.IService, codec types.ICodec) {
	addr := network.engine.AppConf().InteriorAddr
	if addr == "" {
		return
	}
	ser.Init(addr, codec, network.engine, network.context)
	network.interior = ser
}
