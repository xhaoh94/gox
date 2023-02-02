package types

type (
	//IServiceDiscovery 服务发现接口
	IServiceDiscovery interface {
		// GetServiceEntityByID 通过id获取服务配置
		GetServiceEntityByID(uint) IServiceEntity
		// GetServiceEntitysByType 获取对应类型的所有服务配置
		GetServiceEntitysByType(string) []IServiceEntity
	}
	//IServiceEntity 服务器配置
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
