package util

import (
	"reflect"
)

func ToCmd(req interface{}, rsp interface{}) uint32 {

	var key string
	if req != nil {
		reqT := reflect.TypeOf(req)
		key = reqT.Elem().Name()
	}
	if rsp != nil {
		rspT := reflect.TypeOf(rsp)
		key = key + rspT.Elem().Name()
	}
	if key == "" {
		return 0
	}
	ukey := StringToHash(key)
	return ukey
}
