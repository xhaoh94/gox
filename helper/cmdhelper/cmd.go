package cmdhelper

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/helper/strhelper"
	"github.com/xhaoh94/gox/xlog"
)

var (
	key2ukey map[string]uint32 = make(map[string]uint32)
	mux      sync.RWMutex
)

func ToCmdByRtype(in reflect.Type, out reflect.Type, actorId uint32) uint32 {
	var key string
	if in != nil {
		if in.Kind() != reflect.Ptr {
			xlog.Error("ToCmdByRtype参数需要是指针类型")
			return 0
		}
		key = in.Elem().Name()
	}
	if out != nil {
		if out.Kind() != reflect.Ptr {
			xlog.Error("ToCmdByRtype参数需要是指针类型")
			return 0
		}
		key = key + out.Elem().Name()
	}
	if key == "" {
		return 0
	}
	if actorId > 0 {
		key = strhelper.ValToString(actorId) + key
	}

	mux.RLock()
	uKey, ok := key2ukey[key]
	mux.RUnlock()
	if !ok {
		mux.Lock()
		uKey = strhelper.StringToHash(key)
		key2ukey[key] = uKey
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
