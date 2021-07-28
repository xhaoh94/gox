package types

import (
	"reflect"

	"google.golang.org/grpc"
)

type (
	//IService 服务器接口
	IService interface {
		Init(string, IEngine)
		Start()
		Stop()
		GetAddr() string
		GetSessionByAddr(addr string) ISession
		GetSessionById(sid string) ISession
	}
	//IChannel 信道接口
	IChannel interface {
		Start()
		Stop()
		Send(data []byte)
		RemoteAddr() string
		LocalAddr() string
		SetCallBackFn(func([]byte), func())
	}
	//ISession 会话接口
	ISession interface {
		UID() string
		RemoteAddr() string
		LocalAddr() string
		Send(uint32, interface{})
		Call(interface{}, interface{}) IDefaultRPC
		Reply(interface{}, uint32)
		Actor(uint32, uint32, interface{})
		SendData([]byte)
	}

	IEngine interface {
		GetServiceID() string
		GetServiceType() string
		GetServiceVersion() string
		GetEvent() IEvent
		GetNetWork() INetwork
		GetCodec() ICodec
	}

	INetwork interface {
		GetSessionById(string) ISession
		GetSessionByAddr(string) ISession
		GetGRpcServer() IGrpcServer
		GetActor() IActor
		GetServiceReg() IServiceReg
		GetOutsideAddr() string
		GetInteriorAddr() string
		GetRpcAddr() string
		RegisterType(uint32, reflect.Type)
		GetRegProtoMsg(uint32) interface{}
	}
	IServiceConfig interface {
		GetRpcAddr() string
		GetOutsideAddr() string
		GetInteriorAddr() string
		GetServiceID() string
		GetServiceType() string
		GetVersion() string
	}
	IServiceReg interface {
		GetServiceConfByID(string) IServiceConfig
		GetServiceConfListByType(string) []IServiceConfig
	}
	IGrpcServer interface {
		GetAddr() string
		GetServer() *grpc.Server
		GetConnByAddr(addr string) *grpc.ClientConn
	}
	//IDefaultRPC rpc
	IDefaultRPC interface {
		Await() bool
	}

	IActor interface {
		Register(uint32, string)
		Relay(uint32, []byte)
		Send(uint32, uint32, interface{})
	}

	//IModule 模块定义
	IModule interface {
		Start(IModule, IEngine)
		Destroy(IModule)
		Put(IModule)
		OnInit()
		OnDestroy()
	}

	IEvent interface {
		Bind(event interface{}, task interface{}) error
		Call(event interface{}, params ...interface{}) ([]reflect.Value, error)
		UnBind(event interface{}) error
		UnBinds()
		Has(event interface{}) bool
		Events() []interface{}
		EventCount() int
	}

	ICodec interface {
		Encode(interface{}) ([]byte, error)
		Decode([]byte, interface{}) error
	}
)
