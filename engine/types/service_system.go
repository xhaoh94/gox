package types

type (
	//服务发现系统
	IServiceSystem interface {
		// 通过id获取服务配置
		GetServiceEntityByID(uint) IServiceEntity
		// 获取对应类型的所有服务配置
		GetServiceEntitysByType(string) []IServiceEntity

		// 获取对应类型的所有服务配置
		GetServiceEntitys() []IServiceEntity
	}
	//服务器配置
	IServiceEntity interface {
		GetID() uint
		GetType() string
		GetVersion() string
		//GetRpcAddr 获取rpc地址
		GetRpcAddr() string
		//GetOutsideAddr 获取外部通信地址
		GetOutsideAddr() string
		//GetInteriorAddr 获取内部通信地址
		GetInteriorAddr() string
	}
)