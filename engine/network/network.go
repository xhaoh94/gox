package network

import (
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

		grpcServer *rpc.GRpcServer
		atr        *actor.Actor
		serviceReg *ServiceReg
		msgid2type map[uint32]reflect.Type
	}
)

func New(engine types.IEngine) *NetWork {
	nw := new(NetWork)
	nw.engine = engine
	nw.atr = actor.New(engine)
	nw.serviceReg = newServiceReg(nw)
	nw.msgid2type = make(map[uint32]reflect.Type)
	return nw
}

func (nw *NetWork) GetServiceReg() types.IServiceReg {
	return nw.serviceReg
}

func (nw *NetWork) GetActor() types.IActor {
	return nw.atr
}
func (nw *NetWork) GetGRpcServer() types.IGrpcServer {
	return nw.grpcServer
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
	if nw.grpcServer != nil {
		return nw.grpcServer.GetAddr()
	}
	return ""
}

//RegisterType 注册协议消息体类型
func (nw *NetWork) RegisterType(msgid uint32, protoType reflect.Type) {
	if _, ok := nw.msgid2type[msgid]; ok {
		xlog.Error("message %s is already registered", msgid)
		return
	}
	nw.msgid2type[msgid] = protoType
}

//GetRegProtoMsg 获取协议消息体
func (nw *NetWork) GetRegProtoMsg(msgid uint32) interface{} {
	msgType, ok := nw.msgid2type[msgid]
	if !ok {
		return nil
	}
	return util.TypeToInterface(msgType)
}

func (nw *NetWork) Start() {

	if nw.interior == nil {
		xlog.Fatal("service is nil")
		return
	}
	nw.interior.Start()
	if nw.outside != nil {
		nw.outside.Start()
	}
	if nw.grpcServer != nil {
		nw.grpcServer.Start()
	}
	nw.atr.Start()
	nw.serviceReg.Start()
}
func (nw *NetWork) Stop() {
	nw.serviceReg.Stop()
	nw.atr.Stop()
	if nw.outside != nil {
		nw.outside.Stop()
	}
	if nw.grpcServer != nil {
		nw.grpcServer.Stop()
	}
	nw.interior.Stop()
}

//SetOutsideService 设置外部服务类型
func (nw *NetWork) SetOutsideService(ser types.IService, addr string) {
	nw.outside = ser
	nw.outside.Init(addr, nw.engine)
}

//SetInteriorService 设置内部服务类型
func (nw *NetWork) SetInteriorService(ser types.IService, addr string) {
	nw.interior = ser
	nw.interior.Init(addr, nw.engine)
}

//SetGrpcAddr 设置grpc服务
func (nw *NetWork) SetGrpcAddr(addr string) {
	nw.grpcServer = rpc.NewGrpcServer(addr, nw.engine)
}
