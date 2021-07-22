package types

import (
	"github.com/xhaoh94/gox/engine/network/rpc"
)

type (
	//IService 服务器接口
	IService interface {
		Init(string)
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
		Call(interface{}, interface{}) rpc.IDefaultRPC
		Reply(interface{}, uint32)
		Actor(uint32, uint32, interface{})
		SendData([]byte)
	}
)
