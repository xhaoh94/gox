package types

import "context"

type (
	//模块接口
	IModule interface {
		Init(IModule)
		Start(IModule)
		Destroy(IModule)

		//注册协议或添加子模块写在这里
		OnInit()
		//业务逻辑初始化写这里
		OnStart()
		//模块销毁
		OnDestroy()
	}

	ProtoFn[V any] interface {
		func(context.Context, ISession, V)
	}
	ProtoRPCFn[V1 any, V2 any] interface {
		func(context.Context, ISession, V1) (V2, error)
	}
)
