package network

import (
	"context"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/network/actor"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
)

type (
	NetWork struct {
		engine   types.IEngine
		outside  types.IService
		interior types.IService

		context   context.Context
		contextFn context.CancelFunc

		rpc     *rpc.RPC
		atrCrtl *actor.ActorCrtl
		svCrtl  *ServiceCrtl
		cmdType map[uint32]reflect.Type
		cmdLock sync.RWMutex
	}
)

func New(engine types.IEngine, ctx context.Context) *NetWork {
	nw := new(NetWork)
	nw.engine = engine
	nw.atrCrtl = actor.New(engine)
	nw.rpc = rpc.New(engine)
	nw.svCrtl = newServiceReg(nw)
	nw.cmdType = make(map[uint32]reflect.Type)
	nw.context, nw.contextFn = context.WithCancel(ctx)
	return nw
}

func (nw *NetWork) GetServiceCtrl() types.IServiceCtrl {
	return nw.svCrtl
}
func (nw *NetWork) GetActorCtrl() types.IActorCtrl {
	return nw.atrCrtl
}
func (nw *NetWork) GetRPC() types.IGRPC {
	return nw.rpc
}

//GetSession 通过id获取Session
func (nw *NetWork) GetSessionById(sid uint32) types.ISession {
	session := nw.interior.GetSessionById(sid)
	if session == nil && nw.outside != nil {
		session = nw.outside.GetSessionById(sid)
	}
	return session
}

//GetSessionByAddr 通过地址获取Session
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
func (nw *NetWork) GetRpcAddr() string {
	if nw.rpc != nil {
		return nw.rpc.GetAddr()
	}
	return ""
}

//RegisterRType 注册协议消息体类型
func (nw *NetWork) RegisterRType(cmd uint32, protoType reflect.Type) {
	defer nw.cmdLock.Unlock()
	nw.cmdLock.Lock()
	if _, ok := nw.cmdType[cmd]; ok {
		xlog.Error("重复注册协议 cmd[%s]", cmd)
		return
	}
	nw.cmdType[cmd] = protoType
}

//RegisterRType 注册协议消息体类型
func (nw *NetWork) UnRegisterRType(cmd uint32) {
	defer nw.cmdLock.Unlock()
	nw.cmdLock.Lock()
	if _, ok := nw.cmdType[cmd]; ok {
		delete(nw.cmdType, cmd)
	}
}

//GetRegProtoMsg 获取协议消息体
func (nw *NetWork) GetRegProtoMsg(cmd uint32) interface{} {
	nw.cmdLock.RLock()
	rType, ok := nw.cmdType[cmd]
	nw.cmdLock.RUnlock()
	if !ok {
		return nil
	}
	return util.RTypeToInterface(rType)
}

func (nw *NetWork) Start() {

	if nw.interior == nil {
		xlog.Fatal("没有初始化内部网络通信")
		return
	}
	nw.interior.Start()
	if nw.outside != nil {
		nw.outside.Start()
	}
	if nw.rpc != nil {
		nw.rpc.Start()
	}
	nw.svCrtl.Start(nw.context)
	nw.atrCrtl.Start(nw.context)
}
func (nw *NetWork) Stop() {
	nw.contextFn()
	nw.atrCrtl.Stop()
	nw.svCrtl.Stop()
	if nw.outside != nil {
		nw.outside.Stop()
	}
	if nw.rpc != nil {
		nw.rpc.Stop()
	}
	nw.interior.Stop()
}

//SetOutsideService 设置外部服务类型
func (nw *NetWork) SetOutsideService(ser types.IService, addr string) {
	nw.outside = ser
	nw.outside.Init(addr, nw.engine, nw.context)
}

//SetInteriorService 设置内部服务类型
func (nw *NetWork) SetInteriorService(ser types.IService, addr string) {
	nw.interior = ser
	nw.interior.Init(addr, nw.engine, nw.context)
}

//SetGrpcAddr 设置grpc服务
func (nw *NetWork) SetGrpcAddr(addr string) {
	nw.rpc.Init(addr)
}
