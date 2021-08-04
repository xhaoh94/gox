package types

import (
	"context"
	"reflect"

	"google.golang.org/grpc"
)

type (
	//IEngine 引擎接口
	IEngine interface {
		GetServiceID() string
		GetServiceType() string
		GetServiceVersion() string
		GetEvent() IEvent
		GetNetWork() INetwork
		GetCodec() ICodec
	}

	//IModule 模块接口
	IModule interface {
		Start(IModule, IEngine)
		Destroy(IModule)
		Put(IModule)
		OnInit()
		OnDestroy()
	}
	//IEvent 事件接口
	IEvent interface {
		Bind(event interface{}, task interface{}) error
		Call(event interface{}, params ...interface{}) ([]reflect.Value, error)
		UnBind(event interface{}) error
		UnBinds()
		Has(event interface{}) bool
		Events() []interface{}
		EventCount() int
	}
	//INetwork 网络接口
	INetwork interface {
		GetSessionById(string) ISession
		GetSessionByAddr(string) ISession
		GetRPC() IGRPC
		GetActor() IActor
		GetServiceReg() IServiceReg
		GetOutsideAddr() string
		GetInteriorAddr() string
		GetRpcAddr() string
		RegisterRType(uint32, reflect.Type)
		GetRegProtoMsg(uint32) interface{}
	}

	//IService 服务器接口
	IService interface {
		Init(string, IEngine, context.Context)
		Start()
		Stop()
		GetAddr() string
		GetSessionByAddr(string) ISession
		GetSessionById(string) ISession
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
		ID() string
		RemoteAddr() string
		LocalAddr() string
		Send(uint32, interface{}) bool
		Call(interface{}, interface{}) IDefaultRPC
		Reply(interface{}, uint32) bool
		Actor(uint32, uint32, interface{}) bool
	}

	//IServiceReg 服务发现接口
	IServiceReg interface {
		GetServiceConfByID(string) IServiceConfig
		GetServiceConfListByType(string) []IServiceConfig
	}

	//IServiceConfig 服务器配置
	IServiceConfig interface {
		GetRpcAddr() string
		GetOutsideAddr() string
		GetInteriorAddr() string
		GetServiceID() string
		GetServiceType() string
		GetVersion() string
	}

	//IGRPC rpc接口
	IGRPC interface {
		GetAddr() string
		GetServer() *grpc.Server
		GetConnByAddr(string) *grpc.ClientConn
	}
	//IDefaultRPC 内部rpc
	IDefaultRPC interface {
		Await() bool
	}

	//IActor actor接口
	IActor interface {
		Register(uint32, string)
		Send(uint32, uint32, interface{}) bool
	}

	//ICodec 解码编码接口
	ICodec interface {
		Encode(interface{}) ([]byte, error)
		Decode([]byte, interface{}) error
	}
)
