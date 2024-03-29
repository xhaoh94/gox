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
		Outside() IService
		Interior() IService
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
		Codec() ICodec
		Start()
		Stop()
		GetAddr() string
		GetSessionByAddr(string) ISession
		GetSessionById(uint32) ISession
		LinstenByDelSession(callback func(uint32))
	}
	//会话接口
	ISession interface {
		ID() uint32
		Codec(uint32) ICodec
		RemoteAddr() string
		LocalAddr() string
		//发送数据
		Send(uint32, interface{}) bool
		//阻塞等待发送
		Call(interface{}, interface{}) error
		//阻塞等待发送
		CallByCmd(uint32, interface{}, interface{}) error
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
	IRpx interface {
		Await() error
	}
)
