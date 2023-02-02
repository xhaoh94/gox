package types

import (
	"context"
	"encoding/binary"

	"github.com/xhaoh94/gox/engine/app"
)

type (
	//IEngine 引擎接口
	IEngine interface {
		Context() context.Context
		//AppConf 配置
		AppConf() app.AppConf
		//Event 服务事件系统
		Event() IEvent
		//NetWork 网络系统
		NetWork() INetwork
		//Endian 网络大小端
		Endian() binary.ByteOrder
	}
)
