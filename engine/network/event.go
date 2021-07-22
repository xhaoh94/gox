package network

import (
	"reflect"

	"github.com/xhaoh94/gox/engine/event"
	"github.com/xhaoh94/gox/engine/network/proto"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/util"
)

//Register 注册协议对应消息体和回调函数
func Register(msgid uint32, fn interface{}) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	switch tFun.NumIn() {
	case 3:
		in := tFun.In(2)
		if in.Kind() != reflect.Ptr {
			xlog.Error("Register func in != ptr ")
			return
		}
		proto.RegisterType(msgid, in)
		break
	case 2:
		break
	default:
		xlog.Error("Register func parame count fail")
		return
	}
	event.BindNet(msgid, fn)
}

func TestReg(fn interface{}) interface{} {
	return nil
}

//RegisterRPC 注册rpc
func RegisterRPC(args ...interface{}) {
	l := len(args)
	var msgid uint32
	var fn interface{}
	switch l {
	case 1:
		fn = args[0]
		break
	case 2:
		msgid = uint32(args[0].(int))
		fn = args[1]
		break
	default:
		xlog.Error("RegisterRPC parame count fail")
		return
	}
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	if out.Kind() != reflect.Ptr {
		xlog.Error("RegisterRPC func out != ptr ")
		return
	}
	if tFun.NumIn() == 2 {
		in := tFun.In(1)
		if in.Kind() != reflect.Ptr {
			xlog.Error("RegisterRPC func in != ptr ")
			return
		}
		if msgid == 0 {
			key := in.Elem().Name() + out.Elem().Name()
			msgid = util.StrToHash(key)
		}
		proto.RegisterType(msgid, in)
		event.BindNet(msgid, fn)
		return
	}
	if msgid == 0 {
		key := out.Elem().Name()
		msgid = util.StrToHash(key)
	}
	event.BindNet(msgid, fn)
}
