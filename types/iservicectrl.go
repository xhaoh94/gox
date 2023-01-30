package types

type (
	//IServiceDiscovery 服务发现接口
	IServiceDiscovery interface {
		GetServiceConfByID(uint) IServiceEntity
		GetServiceConfListByType(string) []IServiceEntity
	}
	//IServiceEntity 服务器配置
	IServiceEntity interface {
		//GetRpcAddr 获取rpc地址
		GetRpcAddr() string
		//GetOutsideAddr 获取外部通信地址
		GetOutsideAddr() string
		//GetInteriorAddr 获取内部通信地址
		GetInteriorAddr() string
		GetServiceID() uint
		GetServiceType() string
		GetVersion() string
	}
)
