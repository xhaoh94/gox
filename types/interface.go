package types

import (
	"context"
	"encoding/binary"
	"reflect"

	"google.golang.org/grpc"
)

type (
	//IEngine 引擎接口
	IEngine interface {
		ServiceID() uint
		ServiceType() string
		Version() string
		GetEvent() IEvent
		GetNetWork() INetwork
		GetRPC() IGRPC
		GetCodec() ICodec
		GetEndian() binary.ByteOrder
	}

	//IModule 模块接口
	IModule interface {
		Start(IModule, IEngine)
		Stop(IModule)

		OnStart()
		OnStop()
	}
	//IEvent 事件接口
	IEvent interface {
		On(event interface{}, task interface{})
		Off(event interface{}, task interface{})
		Offs(event interface{})
		Has(event interface{}, task interface{}) bool
		Run(event interface{}, params ...interface{})

		Bind(event interface{}, task interface{}) error
		UnBind(event interface{}) error
		UnBinds()
		HasBind(event interface{}) bool
		BindCount() int
		Call(event interface{}, params ...interface{}) ([]reflect.Value, error)
	}

	//INetwork 网络接口
	INetwork interface {
		GetSessionById(uint32) ISession
		GetSessionByAddr(string) ISession
		GetActorCtrl() IActorCtrl
		GetServiceCtrl() IServiceCtrl
		GetOutsideAddr() string
		GetInteriorAddr() string
		RegisterRType(uint32, reflect.Type)
		UnRegisterRType(uint32)
		GetRegProtoMsg(uint32) interface{}
	}
	//IService 服务器接口
	IService interface {
		Init(string, IEngine, context.Context)
		Start()
		Stop()
		GetAddr() string
		GetSessionByAddr(string) ISession
		GetSessionById(uint32) ISession
	}
	//ISession 会话接口
	ISession interface {
		ID() uint32
		RemoteAddr() string
		LocalAddr() string
		Send(uint32, interface{}) bool
		Call(interface{}, interface{}) IDefaultRPC
		Close()
	}
	//IChannel 信道接口
	IChannel interface {
		Start()
		Stop()
		Send(data []byte)
		RemoteAddr() string
		LocalAddr() string
		// SetCallBackFn(func([]byte), func())
		SetSession(ISession)
	}

	//IServiceCtrl 服务发现接口
	IServiceCtrl interface {
		GetServiceConfByID(uint) IServiceConfig
		GetServiceConfListByType(string) []IServiceConfig
	}
	//IServiceConfig 服务器配置
	IServiceConfig interface {
		GetRpcAddr() string
		GetOutsideAddr() string
		GetInteriorAddr() string
		GetServiceID() uint
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

	//IActorCtrl actor发现接口
	IActorCtrl interface {
		Add(IActor)
		Del(IActor)
		Send(uint32, interface{}) bool
		Call(uint32, interface{}, interface{}) IDefaultRPC
	}
	IActor interface {
		AddActorFn(fn interface{})
	}

	//ICodec 解码编码接口
	ICodec interface {
		Encode(interface{}) ([]byte, error)
		Decode([]byte, interface{}) error
	}
)
