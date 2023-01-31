package network

import (
	"context"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/helper/commonhelper"
	"github.com/xhaoh94/gox/network/rpc"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/xlog"
)

type (
	NetWork struct {
		context   context.Context
		contextFn context.CancelFunc

		outside          types.IService
		interior         types.IService
		rpc              types.IRPC
		serviceDiscovery *ServiceDiscovery
		attorDiscovery   *ActorDiscovery
		cmdType          map[uint32]reflect.Type
		cmdLock          sync.RWMutex
	}
)

func New(ctx context.Context, sid uint, sType string, version string) *NetWork {
	network := new(NetWork)
	network.cmdType = make(map[uint32]reflect.Type)
	network.context, network.contextFn = context.WithCancel(ctx)
	network.rpc = rpc.New()
	network.serviceDiscovery = NewServiceDiscovery(network.context, ServiceEntity{
		ServiceID:    sid,
		ServiceType:  sType,
		Version:      version,
		OutsideAddr:  network.OutsideAddr(),
		InteriorAddr: network.InteriorAddr(),
		RPCAddr:      network.Rpc().GetAddr(),
	})

	network.attorDiscovery = NewActorDiscovery(network.context, sid, network)
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
func (network *NetWork) ServiceDiscovery() types.IServiceDiscovery {
	return network.serviceDiscovery
}
func (network *NetWork) ActorDiscovery() types.IActorDiscovery {
	return network.attorDiscovery
}
func (network *NetWork) OutsideAddr() string {
	if network.outside != nil {
		return network.outside.GetAddr()
	}
	return ""
}
func (network *NetWork) InteriorAddr() string {
	if network.interior != nil {
		return network.interior.GetAddr()
	}
	return ""
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
	network.rpc.Start()
	network.serviceDiscovery.Start()
	network.attorDiscovery.Start()
}
func (network *NetWork) Destroy() {
	network.contextFn()
	if network.outside != nil {
		network.outside.Stop()
	}
	network.interior.Stop()
	network.rpc.Stop()
	network.serviceDiscovery.Stop()
	network.attorDiscovery.Stop()
}
func (network *NetWork) Serve() {
	network.rpc.Serve()
}

// SetOutsideService 设置外部服务类型
func (network *NetWork) SetOutsideService(ser types.IService, addr string, codec types.ICodec) {
	if addr == "" {
		return
	}
	network.outside = ser
	network.outside.Init(addr, codec, network.context)
}

// SetInteriorService 设置内部服务类型
func (network *NetWork) SetInteriorService(ser types.IService, addr string, codec types.ICodec) {
	if addr == "" {
		return
	}
	network.interior = ser
	network.interior.Init(addr, codec, network.context)
}

// SetGrpcAddr 设置grpc服务
func (network *NetWork) SetGrpcAddr(addr string) {
	network.rpc.SetAddr(addr)
}
