package proto

import (
	"reflect"

	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/util"
)

var (
	msgid2type map[uint32]reflect.Type = make(map[uint32]reflect.Type)
)

//RegisterType 注册消息体类型
func RegisterType(msgid uint32, protoType reflect.Type) {
	if _, ok := msgid2type[msgid]; ok {
		xlog.Error("message %s is already registered", msgid)
		return
	}
	msgid2type[msgid] = protoType
}

//GetMsg 获取消息体
func GetMsg(msgid uint32) interface{} {
	msgType, ok := msgid2type[msgid]
	if !ok {
		return nil
	}
	return util.TypeToInterface(msgType)
}
