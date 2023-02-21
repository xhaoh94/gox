package cmdhelper

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/xlog"
)

var (
	keyToCmd map[string]uint32 = make(map[string]uint32)
	mux      sync.RWMutex
)

func ToCmdByRtype(in reflect.Type, out reflect.Type, locationID uint32) uint32 {
	var key string
	if in != nil {
		if in.Kind() != reflect.Ptr {
			xlog.Error("ToCmdByRtype:参数需要是指针类型")
			return 0
		}
		key = in.Elem().Name()
	}
	if out != nil {
		if out.Kind() != reflect.Ptr {
			xlog.Error("ToCmdByRtype:参数需要是指针类型")
			return 0
		}
		key = key + out.Elem().Name()
	}
	if key == "" {
		return 0
	}
	if locationID > 0 {
		key = strhelper.ValToString(locationID) + key
	}

	mux.RLock()
	uKey, ok := keyToCmd[key]
	mux.RUnlock()
	if !ok {
		uKey = strhelper.StringToHash(key)
		mux.Lock()
		keyToCmd[key] = uKey
		mux.Unlock()
	}
	return uKey
}
func ToCmd(in interface{}, out interface{}, actorId uint32) uint32 {

	var reqT, rspT reflect.Type
	if in != nil {
		reqT = reflect.TypeOf(in)
	}
	if out != nil {
		rspT = reflect.TypeOf(out)
	}
	return ToCmdByRtype(reqT, rspT, actorId)
}

// 触发
func CallEvt(event uint32, params ...any) (any, error) {
	values, err := gox.Event.Call(event, params...)
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
		return nil, consts.Error_3
	}
}
