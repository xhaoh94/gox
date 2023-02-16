package types

type (
	//定位系统
	ILocationSystem interface {
		Add(ILocationEntity)
		Del(ILocationEntity)
		Send(uint32, interface{}) bool
		Call(uint32, interface{}, interface{}) IRpcx
	}
	ILocationEntity interface {
		ActorID() uint32
		AddActorFn(fn interface{})
		GetFnList() []interface{}
		GetCmdList() []uint32
		SetCmdList(cmd uint32)
		Destroy()
	}
)
