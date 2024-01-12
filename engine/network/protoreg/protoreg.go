package protoreg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"
)

var (
	bingLock  sync.RWMutex
	bingFnMap map[uint32]reflect.Value = make(map[uint32]reflect.Value)

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

// 绑定事件，一个事件只能绑定一个回调，回调可带返回参数
func bind(cmd uint32, task interface{}) error {
	bingLock.Lock()
	defer bingLock.Unlock()
	if _, ok := bingFnMap[cmd]; ok {
		logger.Error().Interface("CMD", cmd).Interface("Task", task).Msg("重复监听事件")
		return fmt.Errorf("protoreg.Bind 重复监听事件 cmd:[%d]", cmd)
	}
	f := reflect.ValueOf(task)
	if f.Type().Kind() != reflect.Func {
		logger.Error().Interface("CMD", cmd).Interface("Task", task).Msg("监听事件对象不是方法")
		return fmt.Errorf("监听事件对象不是方法 CMD:[%v]", cmd)
	}
	bingFnMap[cmd] = f
	return nil
}
func unBind(cmd uint32) error {
	bingLock.Lock()
	defer bingLock.Unlock()
	if _, ok := bingFnMap[cmd]; !ok {
		tip := fmt.Sprintf("没有找到监听的事件 CMD:[%v]", cmd)
		return errors.New(tip)
	}
	delete(bingFnMap, cmd)
	return nil
}

// 触发
func Call(event uint32, ctx context.Context, session types.ISession, require any) (any, error) {
	values, err := call(event, ctx, session, require)
	if err != nil {
		return nil, err
	}
	switch len(values) {
	case 0:
		return nil, nil
	case 1:
		return values[0].Interface(), nil
	case 2:
		var v1 = values[1]
		if v1.IsNil() {
			return values[0].Interface(), nil
		}
		typeOfError := reflect.TypeOf((*error)(nil)).Elem()
		if v1.Type().Implements(typeOfError) {
			return values[0].Interface(), v1.Interface().(error)
		}
		return values[0].Interface(), nil
	default:
		return nil, errors.New("返回参数错误")
	}
}

// 发送事件，存在返回参数
func call(cmd uint32, ctx context.Context, session types.ISession, require any) ([]reflect.Value, error) {
	bingLock.RLock()
	fn, ok := bingFnMap[cmd]
	bingLock.RUnlock()
	if !ok {
		return nil, errors.New("没有找到监听的事件")
	}
	in := make([]reflect.Value, 3)
	in[0] = reflect.ValueOf(ctx)
	in[1] = reflect.ValueOf(session)
	in[2] = reflect.ValueOf(require)
	// numIn := fn.Type().NumIn()
	// for i := range params {
	// 	if i >= numIn {
	// 		break
	// 	}
	// 	param := params[i]
	// 	in[i] = reflect.ValueOf(param)
	// }
	return fn.Call(in), nil
}

func HasBind(cmd uint32) bool {
	bingLock.RLock()
	defer bingLock.RUnlock()
	_, ok := bingFnMap[cmd]
	return ok
}

// 注册协议对应消息体和回调函数
func Register[T types.ProtoFn[V], V any](cmd uint32, fn T) {

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
	bind(cmd, fn)
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
	bind(cmd, fn)
}

// 注册RPC消息
func RegisterRpc[T func(context.Context, types.ISession, V1) (V2, error), V1 any, V2 any](fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	in := tFun.In(2)
	cmd := cmdhelper.ToCmdByRtype(in, out, 0)
	registerRType(cmd, in)
	bind(cmd, fn)
}

// 注册定位消息
func AddLocation[T func(context.Context, types.ISession, V), V any](entity types.ILocation, fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	in := tFun.In(2)
	locationID := entity.LocationID()
	cmd := cmdhelper.ToCmdByRtype(in, nil, locationID)
	registerRType(cmd, in)
	bind(cmd, fn)

	defer locationLock.Unlock()
	locationLock.Lock()
	if _, ok := locationToCmds[locationID]; !ok {
		locationToCmds[locationID] = make([]uint32, 0)
	}
	locationToCmds[locationID] = append(locationToCmds[locationID], cmd)
}

// 注册定位RPC消息
func AddLocationRpc[T func(context.Context, types.ISession, V1) (V2, error), V1 any, V2 any](entity types.ILocation, fn T) {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	out := tFun.Out(0)
	in := tFun.In(2)
	locationID := entity.LocationID()
	cmd := cmdhelper.ToCmdByRtype(in, out, locationID)
	registerRType(cmd, in)
	bind(cmd, fn)

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
func RemoveLocation(entity types.ILocation) {
	defer locationLock.Unlock()
	locationLock.Lock()
	locationID := entity.LocationID()
	if cmdList, ok := locationToCmds[locationID]; ok {
		for index := range cmdList {
			cmd := cmdList[index]
			unRegisterRType(cmd)
			unBind(cmd)
		}
	}
}
