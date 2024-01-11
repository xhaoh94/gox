package types

type (
	//定位系统
	ILocationSystem interface {
		//添加实体
		Add(ILocation)
		//添加实体列表
		Adds([]ILocation)
		//删除实体
		Del(ILocation)
		//删除实体列表
		Dels([]ILocation)
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
