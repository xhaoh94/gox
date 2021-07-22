package module

import "sync"

type (
	//IModule 模块定义
	IModule interface {
		Init(self IModule)
		Destroy(self IModule)
		Put(mod IModule)
		OnInit()
		OnDestroy()
	}

	//Module 模块
	Module struct {
		childModules []IModule
		lock         sync.Mutex
	}
)

//Init 初始化模块
func (m *Module) Init(self IModule) {
	self.OnInit()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Init(v)
		}
	}
}

//Put 添加模块
func (m *Module) Put(mod IModule) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if m.childModules == nil {
		m.childModules = make([]IModule, 0)
	}
	m.childModules = append(m.childModules, mod)
}

//Destroy 销毁模块
func (m *Module) Destroy(self IModule) {
	for i := range m.childModules {
		v := m.childModules[i]
		v.Destroy(v)
	}
	self.OnDestroy()
}

//OnDestroy 进行销毁
func (m *Module) OnDestroy() {

}
