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
		//获取进程间通信Session
		GetSessionByAppID(uint) ISession
		//Rpc
		Rpc() IRPC
		// 通过id获取服务配置
		GetServiceEntityByID(uint) IServiceEntity
		// 获取对应类型的所有服务配置
		GetServiceEntitys(...ServiceOptionFunc) []IServiceEntity
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
		CallByCmd(uint32, interface{}, interface{}) IRpcx
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
		//获取GRpc服务
		GRpcServer() *grpc.Server
		//通过地址获取GRPC客户端
		GetClientConnByAddr(string) *grpc.ClientConn
	}
	//内部rpc
	IRpcx interface {
		Await() error
	}
)
