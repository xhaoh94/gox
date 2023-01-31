package types

type (
	IDiscovery interface {
		Init()
		Destroy()
		//Actor 获取Actor管理器
		Actor() IActorDiscovery
		//Service 获取服务注册发现管理器
		Service() IServiceDiscovery
	}
)
