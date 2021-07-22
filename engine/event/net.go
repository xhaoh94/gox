package event

import (
	"github.com/xhaoh94/gox/consts"
)

var netEvt = New()

//BindNet 绑定网络消息事件
func BindNet(event uint32, task interface{}) error {
	return netEvt.bind(event, task)
}

//CallNet 触发
func CallNet(event uint32, params ...interface{}) (interface{}, error) {
	values, err := netEvt.call(event, params...)
	if err != nil {
		return nil, err
	}
	switch len(values) {
	case 0:
		return nil, nil
	case 1:
		return values[0].Interface(), nil
	case 2:
		return values[0].Interface(), (values[1].Interface()).(error)
	default:
		return nil, consts.CallNetError
	}
}
