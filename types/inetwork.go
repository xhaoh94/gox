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
		//OutsideAddr 获得外部通信地址
		OutsideAddr() string
		//InteriorAddr 获得内部通信地址
		InteriorAddr() string
		//Rpc rpc系统
		Rpc() IRPC
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
		Call(interface{}, interface{}) IXRPC
		//ActorCall
		ActorCall(uint32, interface{}, interface{}) IXRPC
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
		Start()
		Stop()
		Serve()
		//SetAddr 设置rpc地址
		SetAddr(string)
		//GetAddr 获取rpc地址
		GetAddr() string
		//GetServer 获取GRpc服务
		GetServer() *grpc.Server
		//GetConnByAddr 通过地址获取GRPC客户端
		GetConnByAddr(string) *grpc.ClientConn
		Put(IXRPC)
		ParseMsg(IByteArray, ICodec)
	}
	//IXRPC 内部rpc
	IXRPC interface {
		Await() bool
		RID() uint32
	}
)
