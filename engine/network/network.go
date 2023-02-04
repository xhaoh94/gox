package network

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	NetWork struct {
		__init        bool
		outside       types.IService
		interior      types.IService
		rpc           types.IRPC
		serviceSystem types.IServiceSystem
		actorSystem   types.IActorSystem
		cmdType       map[uint32]reflect.Type
		cmdLock       sync.RWMutex
	}
)

func New() *NetWork {
	network := new(NetWork)
	network.cmdType = make(map[uint32]reflect.Type)
	network.rpc = rpc.New()
	network.actorSystem = newActorSystem(gox.Ctx)
	network.serviceSystem = newServiceSystem(gox.Ctx)
	return network
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
func (network *NetWork) ServiceSystem() types.IServiceSystem {
	return network.serviceSystem
}
func (network *NetWork) ActorSystem() types.IActorSystem {
	return network.actorSystem
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
		xlog.Fatal("网络系统: 需要设置InteriorService")
		return
	}
	if network.__init {
		xlog.Error("网络系统: 重复初始化")
		return
	}
	network.__init = true
	network.interior.Start()
	if network.outside != nil {
		network.outside.Start()
	}
	network.rpc.(*rpc.RPC).Start()
	network.serviceSystem.(*ServiceSystem).Start()
	network.actorSystem.(*ActorSystem).Start()
}
func (network *NetWork) Destroy() {
	if !network.__init {
		return
	}
	network.__init = false

	if network.outside != nil {
		network.outside.Stop()
	}
	network.interior.Stop()
	network.rpc.(*rpc.RPC).Stop()
	network.serviceSystem.(*ServiceSystem).Stop()
	network.actorSystem.(*ActorSystem).Stop()
}
func (network *NetWork) Serve() {
	network.rpc.(*rpc.RPC).Serve()
}

// SetOutsideService 设置外部服务类型
func (network *NetWork) SetOutsideService(ser types.IService, codec types.ICodec) {
	addr := gox.AppConf.OutsideAddr
	if addr == "" {
		return
	}
	ser.Init(addr, codec)
	network.outside = ser
}

// SetInteriorService 设置内部服务类型
func (network *NetWork) SetInteriorService(ser types.IService, codec types.ICodec) {
	addr := gox.AppConf.InteriorAddr
	if addr == "" {
		return
	}
	ser.Init(addr, codec)
	network.interior = ser
}
