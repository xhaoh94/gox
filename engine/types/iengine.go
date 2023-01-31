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
		//NetWork 网络系统
		NetWork() INetwork
		//Endian 网络大小端
		Endian() binary.ByteOrder
	}
)
