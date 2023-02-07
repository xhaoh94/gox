package protoreg

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/xlog"
)

var (
	cmdType map[uint32]reflect.Type = make(map[uint32]reflect.Type)
	cmdLock sync.RWMutex
)

// 注册协议消息体类型
func RegisterRType(cmd uint32, protoType reflect.Type) {
	defer cmdLock.Unlock()
	cmdLock.Lock()
	if _, ok := cmdType[cmd]; ok {
		xlog.Error("重复注册协议 cmd[%s]", cmd)
		return
	}
	cmdType[cmd] = protoType
}

// 注销协议消息体类型
func UnRegisterRType(cmd uint32) {
	defer cmdLock.Unlock()
	cmdLock.Lock()
	delete(cmdType, cmd)
}

// 获取协议消息体
func GetProtoMsg(cmd uint32) interface{} {
	cmdLock.RLock()
	rType, ok := cmdType[cmd]
	cmdLock.RUnlock()
	if !ok {
		return nil
	}
	return commonhelper.RTypeToInterface(rType)
}
