package commonhelper

import (
	"reflect"

	"github.com/google/uuid"
)

// 获取唯一id
func NewUUID() string {
	id := uuid.New()
	return id.String()
}

// 通过类型获取实体
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

type Number interface {
	int | uint | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func DeleteSlice[T Number | string](list []T, elem T) []T {
	j := 0
	for _, v := range list {
		if v != elem {
			list[j] = v
			j++
		}
	}
	return list[:j]
}
