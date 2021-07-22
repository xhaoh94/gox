package event

import "reflect"

var defaultEvent = New()

//Bind 绑定事件
func Bind(event string, task interface{}) error {
	return defaultEvent.bind(event, task)
}

//Call 触发
func Call(event string, params ...interface{}) ([]reflect.Value, error) {
	return defaultEvent.call(event, params...)
}

//UnBind 取消绑定
func UnBind(event string) error {
	return defaultEvent.unBind(event)
}

//UnBinds 取消所有事件
func UnBinds() {
	defaultEvent.unBinds()
}

//Has 是否有这个事件
func Has(event string) bool {
	return defaultEvent.hasEvent(event)
}
