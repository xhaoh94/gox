package types

type (
	//服务器配置
	IServiceEntity interface {
		GetID() uint
		GetType() string
		GetVersion() string
		IsLocation() bool
		//GetRpcAddr 获取rpc地址
		GetRpcAddr() string
		//GetOutsideAddr 获取外部通信地址
		GetOutsideAddr() string
		//GetInteriorAddr 获取内部通信地址
		GetInteriorAddr() string
	}
	ServiceOptionFunc func(entity IServiceEntity) bool
)

// 获取指定类型的服务器
func WithType(t string) ServiceOptionFunc {
	return func(entity IServiceEntity) bool {
		return entity.GetType() == t
	}
}

// 获取开启定位的服务器
func WithLocation() ServiceOptionFunc {
	return func(entity IServiceEntity) bool {
		return entity.IsLocation()
	}
}

// 获取指定ID外的服务器
func WithExcludeID(id uint) ServiceOptionFunc {
	return func(entity IServiceEntity) bool {
		return entity.GetID() != id
	}
}

// 获取指定ID外的服务器
func WithExcludeIDs(ids []uint) ServiceOptionFunc {
	return func(entity IServiceEntity) bool {
		if len(ids) == 0 {
			return true
		}
		for _, id := range ids {
			if entity.GetID() == id {
				return false
			}
		}
		return true
	}
}
