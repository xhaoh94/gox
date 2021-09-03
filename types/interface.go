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
		//GetEvent 服务事件系统
		GetEvent() IEvent
		//GetNetWork 网络系统
		GetNetWork() INetwork
		//GetRPC rpc系统
		GetRPC() IGRPC
		//GetCodec 解码系统
		GetCodec() ICodec
		//GetEndian 网络大小端
		GetEndian() binary.ByteOrder
	}

	//IModule 模块接口
	IModule interface {
		Init(IModule, IEngine)
		Destroy(IModule)

		//注册协议或添加子模块写在这里
		OnInit()
		//业务逻辑初始化写这里
		OnStart()
		//模块销毁
		OnDestroy()
	}
	//IEvent 事件接口
	IEvent interface {
		//On 事件监听
		On(event interface{}, task interface{})
		//Off 事件取消监听
		Off(event interface{}, task interface{})
		//Offs 取消所有监听源
		Offs(event interface{})
		//Has 是否监听此事件
		Has(event interface{}, task interface{}) bool
		//Run 派发事件
		Run(event interface{}, params ...interface{})

		//Bind 事件绑定，跟on的区别在于。此方法是同步的，且一个event只能对应一个事件。且可带返回值
		Bind(event interface{}, task interface{}) error
		//UnBind 取消事件绑定
		UnBind(event interface{}) error
		//UnBinds取消所有事件绑定
		UnBinds()
		//HasBind 是否拥有事件绑定
		HasBind(event interface{}) bool
		//BindCount 绑定数量
		BindCount() int
		//Call 事件响应
		Call(event interface{}, params ...interface{}) ([]reflect.Value, error)
	}

	//INetwork 网络接口
	INetwork interface {
		//GetSessionById 通过Id获取通信Session
		GetSessionById(uint32) ISession
		//GetSessionByAddr 通过地址获取通信Session
		GetSessionByAddr(string) ISession
		//GetActorCtrl 获取Actor管理器
		GetActorCtrl() IActorCtrl
		//GetServiceCtrl 获取服务注册发现管理器
		GetServiceCtrl() IServiceCtrl
		//GetOutsideAddr 获得外部通信地址
		GetOutsideAddr() string
		//GetInteriorAddr 获得内部通信地址
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
		//Send 发送数据
		Send(uint32, interface{}) bool
		//Call RPC请求
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
		//GetRpcAddr 获取rpc地址
		GetRpcAddr() string
		//GetOutsideAddr 获取外部通信地址
		GetOutsideAddr() string
		//GetInteriorAddr 获取内部通信地址
		GetInteriorAddr() string
		GetServiceID() uint
		GetServiceType() string
		GetVersion() string
	}

	//IGRPC rpc接口
	IGRPC interface {
		//GetAddr 获取rpc地址
		GetAddr() string
		//GetServer 获取GRpc服务
		GetServer() *grpc.Server
		//GetConnByAddr 通过地址获取GRPC客户端
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
		ActorID() uint32
		AddActorFn(fn interface{})
		GetFnList() []interface{}
		GetCmdList() []uint32
		SetCmdList(cmd uint32)
		Destroy()
	}

	//ICodec 解码编码接口
	ICodec interface {
		Encode(interface{}) ([]byte, error)
		Decode([]byte, interface{}) error
	}
)
