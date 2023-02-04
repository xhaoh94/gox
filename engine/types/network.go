package types

import (
	"reflect"

	"google.golang.org/grpc"
)

type (
	//INetwork 网络接口
	INetwork interface {
		Init()
		Destroy()
		//GetSessionById 通过Id获取通信Session
		GetSessionById(uint32) ISession
		//GetSessionByAddr 通过地址获取通信Session
		GetSessionByAddr(string) ISession
		//Rpc
		Rpc() IRPC
		//服务发现
		ServiceSystem() IServiceSystem
		//Actor注册
		ActorSystem() IActorSystem
		RegisterRType(uint32, reflect.Type)
		UnRegisterRType(uint32)
		GetRegProtoMsg(uint32) interface{}
	}
	//IService 服务器接口
	IService interface {
		Init(string, ICodec)
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
		Call(interface{}, interface{}) IRpcx
		//ActorCall
		ActorCall(uint32, interface{}, interface{}) IRpcx
		Close()
	}
	//IChannel 信道接口
	IChannel interface {
		Start()
		Stop()
		Send(data []byte)
		RemoteAddr() string
		LocalAddr() string
		SetSession(ISession)
	}

	//IRPC rpc接口
	IRPC interface {
		Serve()
		//GetServer 获取GRpc服务
		GetServer() *grpc.Server
		//GetClientConnByAddr 通过地址获取GRPC客户端
		GetClientConnByAddr(string) *grpc.ClientConn
	}
	//IRpcx 内部rpc
	IRpcx interface {
		Await() bool
	}
)
