package protoreg

import (
	"context"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"
)

var (
	cmdType        map[uint32]reflect.Type = make(map[uint32]reflect.Type)
	cmdLock        sync.RWMutex
	locationToCmds map[uint32][]uint32
	locationLock   sync.RWMutex
)

// 注册协议消息体类型
func registerRType(cmd uint32, protoType reflect.Type) {
	defer cmdLock.Unlock()
	cmdLock.Lock()
	if _, ok := cmdType[cmd]; ok {
		logger.Error().Uint32("CMD", cmd).Msg("重复注册协议")
		return
	}
	cmdType[cmd] = protoType
}

// 注销协议消息体类型
func unRegisterRType(cmd uint32) {
	defer cmdLock.Unlock()
	cmdLock.Lock()
	delete(cmdType, cmd)
}

// 获取协议消息体
func GetRequireByCmd(cmd uint32) interface{} {
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
		logger.Error().Uint32("CMD", cmd).Msg("协议回调函数不是方法")
		return
	}
	in := tFun.In(2)
	if in.Kind() != reflect.Ptr {
		logger.Error().Interface("In", in).Uint32("CMD", cmd).Msg("协议回调函数参数需要是指针类型")
		return
	}
	registerRType(cmd, in)
	gox.Event.Bind(cmd, fn)
}

// 注册带CMD的RPC消息
func RegisterRpcCmd[T func(context.Context, types.ISession, V1) (V2, error), V1 any, V2 any](cmd uint32, fn T) {

	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	if out.Kind() != reflect.Ptr {
		logger.Error().Interface("Out", out).Msg("RPC函数参数需要是指针类型")
		return
	}

	in := tFun.In(2)
	if in.Kind() != reflect.Ptr {
		logger.Error().Interface("In", in).Msg("RPC函数参数需要是指针类型")
		return
	}
	registerRType(cmd, in)
	gox.Event.Bind(cmd, fn)
}

// 注册RPC消息
func RegisterRpc[T func(context.Context, types.ISession, V1) (V2, error), V1 any, V2 any](fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	in := tFun.In(2)
	cmd := cmdhelper.ToCmdByRtype(in, out, 0)
	registerRType(cmd, in)
	gox.Event.Bind(cmd, fn)
}

// 注册定位消息
func AddLocation[T func(context.Context, types.ISession, V), V any](entity types.ILocationEntity, fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	in := tFun.In(2)
	locationID := entity.LocationID()
	cmd := cmdhelper.ToCmdByRtype(in, nil, locationID)
	registerRType(cmd, in)
	gox.Event.Bind(cmd, fn)

	defer locationLock.Unlock()
	locationLock.Lock()
	if _, ok := locationToCmds[locationID]; !ok {
		locationToCmds[locationID] = make([]uint32, 0)
	}
	locationToCmds[locationID] = append(locationToCmds[locationID], cmd)
}

// 注册定位RPC消息
func AddLocationRpc[T func(context.Context, types.ISession, V1) (V2, error), V1 any, V2 any](entity types.ILocationEntity, fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	in := tFun.In(2)
	locationID := entity.LocationID()
	cmd := cmdhelper.ToCmdByRtype(in, out, locationID)
	registerRType(cmd, in)
	gox.Event.Bind(cmd, fn)

	defer locationLock.Unlock()
	locationLock.Lock()
	if locationToCmds == nil {
		locationToCmds = make(map[uint32][]uint32)
	}
	if _, ok := locationToCmds[locationID]; !ok {
		locationToCmds[locationID] = make([]uint32, 0)
	}
	locationToCmds[locationID] = append(locationToCmds[locationID], cmd)
}

// 注销定位消息
func RemoveLocation(entity types.ILocationEntity) {
	defer locationLock.Unlock()
	locationLock.Lock()
	locationID := entity.LocationID()
	if cmdList, ok := locationToCmds[locationID]; ok {
		for index := range cmdList {
			cmd := cmdList[index]
			unRegisterRType(cmd)
			gox.Event.UnBind(cmd)
		}
	}
}
