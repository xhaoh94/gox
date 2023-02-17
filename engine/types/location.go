package types

type (
	//定位系统
	ILocationSystem interface {
		Add(ILocationEntity)
		Adds([]ILocationEntity)
		Del(ILocationEntity)
		Dels([]ILocationEntity)
	}
	ILocationEntity interface {
		Init(ILocationEntity) bool
		OnInit()
		Destroy()
		ActorID() uint32
		Register(fn interface{})
	}
)
