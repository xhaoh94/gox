package network

import (
	"context"
	"reflect"

	"github.com/xhaoh94/gox/engine/network/actor"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/util"
)

type (
	NetWork struct {
		engine   types.IEngine
		outside  types.IService
		interior types.IService

		context   context.Context
		contextFn context.CancelFunc

		rpc        *rpc.RPC
		atr        *actor.Actor
		serviceReg *ServiceReg
		cmd2type   map[uint32]reflect.Type
	}
)

func New(engine types.IEngine, ctx context.Context) *NetWork {
	nw := new(NetWork)
	nw.engine = engine
	nw.atr = actor.New(engine)
	nw.rpc = rpc.New(engine)
	nw.serviceReg = newServiceReg(nw)
	nw.cmd2type = make(map[uint32]reflect.Type)
	nw.context, nw.contextFn = context.WithCancel(ctx)
	return nw
}

func (nw *NetWork) GetServiceReg() types.IServiceReg {
	return nw.serviceReg
}

func (nw *NetWork) GetActor() types.IActor {
	return nw.atr
}
func (nw *NetWork) GetRPC() types.IGRPC {
	return nw.rpc
}

//GetSession 通过id获取Session
func (nw *NetWork) GetSessionById(sid string) types.ISession {
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
	if _, ok := nw.cmd2type[cmd]; ok {
		xlog.Error("重复注册协议 msgid->[%s]", cmd)
		return
	}
	nw.cmd2type[cmd] = protoType
}

//GetRegProtoMsg 获取协议消息体
func (nw *NetWork) GetRegProtoMsg(cmd uint32) interface{} {
	rType, ok := nw.cmd2type[cmd]
	if !ok {
		return nil
	}
	return util.RTypeToInterface(rType)
}

func (nw *NetWork) Start() {

	if nw.interior == nil {
		xlog.Fatal("没有内部网络通信")
		return
	}
	nw.interior.Start()
	if nw.outside != nil {
		nw.outside.Start()
	}
	if nw.rpc != nil {
		nw.rpc.Start()
	}
	nw.serviceReg.Start(nw.context)
	nw.atr.Start(nw.context)
}
func (nw *NetWork) Stop() {
	nw.contextFn()
	nw.atr.Stop()
	nw.serviceReg.Stop()
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
