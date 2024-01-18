package gox

import (
	"sync"

	"github.com/xhaoh94/gox/engine/types"
)

type (
	//Module 模块
	Module struct {
		childModules []types.IModule
		lock         sync.Mutex
	}
)

// Init 初始化模块
func (m *Module) Init(self types.IModule) {
	self.OnInit()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Init(v)
		}
	}
}

// OnInit 初始模块
func (mm *Module) OnInit() {

}

func (m *Module) Start(self types.IModule) {
	self.OnStart()
	if m.childModules != nil {
		for i := range m.childModules {
			v := m.childModules[i]
			v.Start(v)
		}
	}
}

// OnStart 模块启动
func (mm *Module) OnStart() {

}

// Destroy 销毁模块
func (m *Module) Destroy(self types.IModule) {
	for i := range m.childModules {
		v := m.childModules[i]
		v.Destroy(v)
	}
	self.OnDestroy()
}

// OnDestroy 模块销毁
func (mm *Module) OnDestroy() {

}

// Put 添加模块
func (m *Module) Put(mod types.IModule) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if m.childModules == nil {
		m.childModules = make([]types.IModule, 0)
	}
	m.childModules = append(m.childModules, mod)
}
