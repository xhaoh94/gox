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

//RTypeToInterface 通过类型获取实体
func RTypeToInterface(t reflect.Type) interface{} {
	if t != nil {
		if t.Kind() != reflect.Ptr {
			return reflect.New(t).Interface()
		} else {
			return reflect.New(t.Elem()).Interface()
		}
	}
	return nil
}
