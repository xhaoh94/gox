package types

import (
	"context"
	"reflect"

	"google.golang.org/grpc"
)

type (
	//INetwork 网络接口
	INetwork interface {
		Engine() IEngine
		//GetSessionById 通过Id获取通信Session
		GetSessionById(uint32) ISession
		//GetSessionByAddr 通过地址获取通信Session
		GetSessionByAddr(string) ISession
		//ActorDiscovery 获取Actor管理器
		ActorDiscovery() IActorDiscovery
		//ServiceDiscovery 获取服务注册发现管理器
		ServiceDiscovery() IServiceDiscovery
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
		Init(string, ICodec, IEngine, context.Context)
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
		//ActorCall
		ActorCall(uint32, interface{}, interface{}) IDefaultRPC
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
)
