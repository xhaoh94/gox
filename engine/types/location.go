package types

type (
	//定位系统
	ILocationSystem interface {
		Add(ILocationEntity)
		Del(ILocationEntity)
	}
	ILocationEntity interface {
		Init()
		Destroy()
		ActorID() uint32
		AddActorFn(fn interface{})
		GetFnList() []interface{}
		GetCmdList() []uint32
		SetCmdList(cmd uint32)
	}
)
