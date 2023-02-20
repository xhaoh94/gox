package types

type (
	//定位系统
	ILocationSystem interface {
		//添加实体
		Add(ILocationEntity)
		//添加实体列表
		Adds([]ILocationEntity)
		//删除实体
		Del(ILocationEntity)
		//删除实体列表
		Dels([]ILocationEntity)
		//获取对应的进程ID
		GetAppID(uint32) uint
		//获取对应的进程ID列表
		GetAppIDs([]uint32) []uint
		//广播
		Broadcast([]uint32, interface{})
		//指定发送
		Send(uint32, interface{}) bool
		//RPC
		Call(uint32, interface{}, interface{}) IRpcx
	}
	ILocationEntity interface {
		Init(ILocationEntity) bool
		OnInit()
		Destroy()
		//定位ID 每个实体的ID都是唯一的，且不变的
		LocationID() uint32
		Register(fn interface{})
	}
)
