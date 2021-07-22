package service

import (
	"reflect"

	"github.com/xhaoh94/gox/util"
)

func converMsgID(req interface{}, rsp interface{}) uint32 {
	rspT := reflect.TypeOf(rsp)
	key := rspT.Elem().Name()
	if req != nil {
		reqT := reflect.TypeOf(req)
		key = reqT.Elem().Name() + key
	}

	ukey := util.StrToHash(key)
	return ukey
}
