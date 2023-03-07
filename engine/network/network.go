package network

import (
	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	NetWork struct {
		__init        bool
		__start       bool
		outside       types.IService
		interior      types.IService
		rpc           *rpc.RPC
		serviceSystem *ServiceSystem
		location      *location.LocationSystem
	}
)

func New() *NetWork {
	return &NetWork{
		rpc:           rpc.New(),
		serviceSystem: newServiceSystem(gox.Ctx),
		location:      location.New(),
	}
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

// 获取进程Session
func (as *NetWork) GetSessionByAppID(appID uint) types.ISession {
	serviceEntity := as.GetServiceEntityByID(appID)
	if serviceEntity == nil {
		logger.Error().Uint("AppID", appID).Msg("没有找到注册的服务")
		return nil
	}
	session := as.GetSessionByAddr(serviceEntity.GetInteriorAddr())
	if session == nil {
		logger.Error().Str("Session InteriorAddr", serviceEntity.GetInteriorAddr()).Msg("没有找到Session")
		return nil
	}
	return session
}

func (network *NetWork) Rpc() types.IRPC {
	return network.rpc
}

func (network *NetWork) Init() {
	if network.interior == nil {
		logger.Fatal().Msg("网络系统: 需要设置InteriorService")
		return
	}
	if network.__init {
		logger.Fatal().Msg("网络系统: 重复初始化")
		return
	}
	network.__init = true
	network.interior.Start()
	if network.outside != nil {
		network.outside.Start()
	}
	network.rpc.Start()
	network.serviceSystem.Start()
	network.location.Init()
}
func (network *NetWork) Start() {
	if network.__start {
		return
	}
	network.__start = true
	network.rpc.Serve()
	network.location.Start()
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
	network.rpc.Stop()
	network.serviceSystem.Stop()
	network.location.Stop()
}

// 通过id获取服务配置
func (ss *NetWork) GetServiceEntityByID(id uint) types.IServiceEntity {
	return ss.serviceSystem.GetServiceEntityByID(id)
}

// // 获取对应类型的所有服务配置
// func (ss *NetWork) GetServiceEntitysByType(appType string) []types.IServiceEntity {
// 	return ss.serviceSystem.GetServiceEntitysByType(appType)
// }

// 获取对应类型的所有服务配置
func (ss *NetWork) GetServiceEntitys(opts ...types.ServiceOptionFunc) []types.IServiceEntity {
	return ss.serviceSystem.GetServiceEntitys(opts...)
}

// SetOutsideService 设置外部服务类型
func (network *NetWork) SetOutsideService(ser types.IService, codec types.ICodec) {
	addr := gox.Config.OutsideAddr
	if addr == "" {
		return
	}
	ser.Init(addr, codec)
	network.outside = ser
}

// SetInteriorService 设置内部服务类型
func (network *NetWork) SetInteriorService(ser types.IService, codec types.ICodec) {
	addr := gox.Config.InteriorAddr
	if addr == "" {
		return
	}
	ser.Init(addr, codec)
	network.interior = ser
}
