package types

type (
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
	ServiceOptionFunc func(entity IServiceEntity) bool
)

func WithType(t string) ServiceOptionFunc {
	return func(entity IServiceEntity) bool {
		return entity.GetType() == t
	}
}

func WithExcludeID(id uint) ServiceOptionFunc {
	return func(entity IServiceEntity) bool {
		return entity.GetID() != id
	}
}
