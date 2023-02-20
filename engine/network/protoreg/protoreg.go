package protoreg

import (
	"context"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/types"
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

// 注册协议对应消息体和回调函数
func Register[T func(context.Context, types.ISession, V), V any](cmd uint32, fn T) {

	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	if tFun.Kind() != reflect.Func {
		xlog.Error("协议回调函数不是方法 cmd:[%d]", cmd)
		return
	}
	in := tFun.In(2)
	if in.Kind() != reflect.Ptr {
		xlog.Error("协议回调函数参数需要是指针类型 cmd[%d]", cmd)
		return
	}
	RegisterRType(cmd, in)
	gox.Event.Bind(cmd, fn)
}

// 注册带CMD的RPC消息
func RegisterRpcCmd[T func(context.Context, V1) V2, V1 any, V2 any](cmd uint32, fn T) {

	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	if out.Kind() != reflect.Ptr {
		xlog.Error("RPC函数参数需要是指针类型")
		return
	}

	in := tFun.In(1)
	if in.Kind() != reflect.Ptr {
		xlog.Error("RPC函数参数需要是指针类型")
		return
	}
	RegisterRType(cmd, in)
	gox.Event.Bind(cmd, fn)
}

// 注册RPC消息
func RegisterRpc[T func(context.Context, V1) V2, V1 any, V2 any](fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	in := tFun.In(1)
	cmd := cmdhelper.ToCmdByRtype(in, out, 0)
	RegisterRType(cmd, in)
	gox.Event.Bind(cmd, fn)
}
