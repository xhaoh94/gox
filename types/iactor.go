package types

type (
	//IActorDiscovery actor发现接口
	IActorDiscovery interface {
		Add(IActorEntity)
		Del(IActorEntity)
		Send(uint32, interface{}) bool
		Call(uint32, interface{}, interface{}) IXRPC
	}
	IActorEntity interface {
		ActorID() uint32
		AddActorFn(fn interface{})
		GetFnList() []interface{}
		GetCmdList() []uint32
		SetCmdList(cmd uint32)
		Destroy()
	}
)
