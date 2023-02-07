package types

import "reflect"

type (
	//事件接口
	IEvent interface {
		//事件监听
		On(event interface{}, task interface{})
		//事件取消监听
		Off(event interface{}, task interface{})
		//取消所有监听源
		Offs(event interface{})
		//是否监听此事件
		Has(event interface{}, task interface{}) bool
		//派发事件
		Run(event interface{}, params ...interface{})

		//事件绑定，跟on的区别在于。此方法是同步的，且一个event只能对应一个事件。且可带返回值
		Bind(event interface{}, task interface{}) error
		//取消事件绑定
		UnBind(event interface{}) error
		//取消所有事件绑定
		UnBinds()
		//是否拥有事件绑定
		HasBind(event interface{}) bool
		//绑定数量
		BindCount() int
		//事件响应
		Call(event interface{}, params ...interface{}) ([]reflect.Value, error)
	}
)
