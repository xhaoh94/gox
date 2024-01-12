package types

type (
	//定位系统
	ILocationSystem interface {
		//注册实体
		Register(ILocation)
		//注册实体列表
		Registers([]ILocation)
		//注销实体
		UnRegister(ILocation)
		//注销实体列表
		UnRegisters([]ILocation)
		//广播
		Broadcast([]uint32, interface{})
		//发送
		Send(uint32, interface{})
		//阻塞等待发送
		Call(uint32, interface{}, interface{}) error
	}
	ILocation interface {
		//定位ID 每个实体的ID都是唯一的，且不变的
		LocationID() uint32
		Init(ILocation)
		OnInit()
		Destroy(ILocation)
	}
)
