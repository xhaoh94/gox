package types

type (
	//IModule 模块接口
	IModule interface {
		Init(IModule, IEngine, func())
		Destroy(IModule)

		//注册协议或添加子模块写在这里
		OnInit()
		//业务逻辑初始化写这里
		OnStart()
		//模块销毁
		OnDestroy()
	}
)
