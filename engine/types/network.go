package types

import (
	"google.golang.org/grpc"
)

type (
	//网络接口
	INetwork interface {
		Init()
		Start()
		Destroy()
		//通过Id获取通信Session
		GetSessionById(uint32) ISession
		//通过地址获取通信Session
		GetSessionByAddr(string) ISession
		//Rpc
		Rpc() IRPC
		//服务发现
		ServiceSystem() IServiceSystem
		//Actor系统
		ActorSystem() IActorSystem
	}
	//服务器接口
	IService interface {
		Init(string, ICodec)
		Start()
		Stop()
		GetAddr() string
		GetSessionByAddr(string) ISession
		GetSessionById(uint32) ISession
	}
	//会话接口
	ISession interface {
		ID() uint32
		RemoteAddr() string
		LocalAddr() string
		//发送数据
		Send(uint32, interface{}) bool
		//RPC请求
		Call(interface{}, interface{}) IRpcx
		//ActorCall
		ActorCall(uint32, interface{}, interface{}) IRpcx
		Close()
	}
	//信道接口
	IChannel interface {
		Start()
		Stop()
		Send(data []byte)
		RemoteAddr() string
		LocalAddr() string
		SetSession(ISession)
	}

	//rpc接口
	IRPC interface {
		//GRpcServer 获取GRpc服务
		GRpcServer() *grpc.Server
		//GetClientConnByAddr 通过地址获取GRPC客户端
		GetClientConnByAddr(string) *grpc.ClientConn
	}
	//内部rpc
	IRpcx interface {
		Await() bool
	}
	//内部rpc
	IActorx interface {
		Await() []byte
	}
)
