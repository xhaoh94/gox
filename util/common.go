package util

import (
	"reflect"

	"github.com/google/uuid"
)

//GetUUID 获取唯一id
func GetUUID() string {
	id := uuid.New()
	return id.String()
}

//TypeToInterface 通过类型获取实体
func TypeToInterface(t reflect.Type) (msg interface{}) {
	if t != nil {
		if t.Kind() != reflect.Ptr {
			msg = reflect.New(t).Interface()
		} else {
			msg = reflect.New(t.Elem()).Interface()
		}
	}
	return
}
