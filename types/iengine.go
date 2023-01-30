package types

import "encoding/binary"

type (
	//IEngine 引擎接口
	IEngine interface {
		EID() uint
		EType() string
		Version() string
		//Event 服务事件系统
		Event() IEvent
		//GetNetWork 网络系统
		GetNetWork() INetwork
		//GetRPC rpc系统
		GetRPC() IGRPC
		//GetEndian 网络大小端
		GetEndian() binary.ByteOrder
	}
)
