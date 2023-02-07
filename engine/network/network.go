package network

import (
	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	NetWork struct {
		__init        bool
		__start       bool
		outside       types.IService
		interior      types.IService
		rpc           types.IRPC
		serviceSystem types.IServiceSystem
		actorSystem   types.IActorSystem
	}
)

func New() *NetWork {
	network := new(NetWork)
	network.rpc = rpc.New()
	network.actorSystem = newActorSystem(gox.Ctx)
	network.serviceSystem = newServiceSystem(gox.Ctx)
	return network
}

// 通过id获取Session
func (network *NetWork) GetSessionById(sid uint32) types.ISession {
	session := network.interior.GetSessionById(sid)
	if session == nil && network.outside != nil {
		session = network.outside.GetSessionById(sid)
	}
	return session
}

// 通过地址获取Session
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
func (network *NetWork) Start() {
	if network.__start {
		return
	}
	network.__start = true
	network.rpc.(*rpc.RPC).Serve()
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
