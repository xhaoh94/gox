package network

import (
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/location"
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
		location      *location.LocationSystem
	}
)

func New() *NetWork {
	network := new(NetWork)
	network.rpc = rpc.New()
	network.serviceSystem = newServiceSystem(gox.Ctx)
	network.location = location.New(gox.Ctx)
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

// 获取进程Session
func (as *NetWork) GetSessionByAppID(appID uint) types.ISession {
	serviceEntity := as.GetServiceEntityByID(appID)
	if serviceEntity == nil {
		xlog.Error("Actor没有找到服务 ServiceID:[%s]", appID)
		return nil
	}
	session := as.GetSessionByAddr(serviceEntity.GetInteriorAddr())
	if session == nil {
		xlog.Error("Actor没有找到session[%d]", serviceEntity.GetInteriorAddr())
		return nil
	}
	return session
}

func (network *NetWork) Rpc() types.IRPC {
	return network.rpc
}

func (network *NetWork) LocationSystem() types.ILocationSystem {
	return network.location
}

func (as *NetWork) LocationSend(actorID uint32, msg interface{}) bool {
	if actorID == 0 {
		xlog.Error("ActorCall传入ActorID不能为空")
		return false
	}
	as.location.RLockCacel(true)
	defer as.location.RLockCacel(false)
	loopCnt := 0
	cmd := cmdhelper.ToCmd(msg, nil, actorID)
	for {
		loopCnt++
		if loopCnt > 5 {
			return false
		}
		if id := as.location.GetAppId(actorID); id > 0 {
			if session := as.GetSessionByAppID(id); session != nil {
				if id == gox.AppConf.Eid {
					if _, err := cmdhelper.CallEvt(cmd, gox.Ctx, session, msg); err == nil {
						return true
					} else {
						xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
					}
				} else {
					return session.Send(cmd, msg)
				}
			}
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
}
func (as *NetWork) LocationCall(actorID uint32, msg interface{}, response interface{}) types.IRpcx {
	if actorID == 0 {
		xlog.Error("ActorCall传入ActorID不能为空")
		return rpc.NewEmptyRpcx()
	}

	as.location.RLockCacel(true)
	defer as.location.RLockCacel(false)
	loopCnt := 0
	cmd := cmdhelper.ToCmd(msg, response, actorID)
	for {
		loopCnt++
		if loopCnt > 5 {
			return rpc.NewEmptyRpcx()
		}
		if id := as.location.GetAppId(actorID); id > 0 {
			if id == gox.AppConf.Eid {
				if response, err := cmdhelper.CallEvt(cmd, gox.Ctx, msg); err == nil {
					rpcx := rpc.NewRpcx(gox.Ctx, response)
					defer rpcx.Run(true)
					return rpcx
				} else {
					xlog.Warn("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
				}
			} else {
				if session := as.GetSessionByAppID(id); session != nil {
					return session.CallByCmd(cmd, msg, response)
				}
			}
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
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
	network.location.Init()
}
func (network *NetWork) Start() {
	if network.__start {
		return
	}
	network.__start = true
	network.rpc.(*rpc.RPC).Serve()
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
	network.rpc.(*rpc.RPC).Stop()
	network.serviceSystem.(*ServiceSystem).Stop()
	network.location.Stop()
	// network.actorSystem.(*ActorSystem).Stop()
}

// 通过id获取服务配置
func (ss *NetWork) GetServiceEntityByID(id uint) types.IServiceEntity {
	return ss.serviceSystem.GetServiceEntityByID(id)
}

// 获取对应类型的所有服务配置
func (ss *NetWork) GetServiceEntitysByType(appType string) []types.IServiceEntity {
	return ss.serviceSystem.GetServiceEntitysByType(appType)
}

// 获取对应类型的所有服务配置
func (ss *NetWork) GetServiceEntitys() []types.IServiceEntity {
	return ss.serviceSystem.GetServiceEntitys()
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
