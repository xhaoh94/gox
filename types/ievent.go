package types

import "reflect"

type (
	//IEvent 事件接口
	IEvent interface {
		//On 事件监听
		On(event interface{}, task interface{})
		//Off 事件取消监听
		Off(event interface{}, task interface{})
		//Offs 取消所有监听源
		Offs(event interface{})
		//Has 是否监听此事件
		Has(event interface{}, task interface{}) bool
		//Run 派发事件
		Run(event interface{}, params ...interface{})

		//Bind 事件绑定，跟on的区别在于。此方法是同步的，且一个event只能对应一个事件。且可带返回值
		Bind(event interface{}, task interface{}) error
		//UnBind 取消事件绑定
		UnBind(event interface{}) error
		//UnBinds取消所有事件绑定
		UnBinds()
		//HasBind 是否拥有事件绑定
		HasBind(event interface{}) bool
		//BindCount 绑定数量
		BindCount() int
		//Call 事件响应
		Call(event interface{}, params ...interface{}) ([]reflect.Value, error)
	}
)
