package actor

import (
	"sync"
)

type (
	Actor struct {
		fnLock  sync.RWMutex
		fnList  []interface{}
		cmdLock sync.RWMutex
		cmdList []uint32
	}
)

// AddActorFn 添加Actor回调
func (art *Actor) AddActorFn(fn interface{}) {
	defer art.fnLock.Unlock()
	art.fnLock.Lock()
	if art.fnList == nil {
		art.fnList = make([]interface{}, 0)
	}
	art.fnList = append(art.fnList, fn)
}

func (art *Actor) Destroy() {
	art.fnList = nil
	art.cmdList = nil
}

func (art *Actor) GetFnList() []interface{} {
	defer art.fnLock.RUnlock()
	art.fnLock.RLock()
	return art.fnList
}

func (art *Actor) GetCmdList() []uint32 {
	defer art.cmdLock.RUnlock()
	art.cmdLock.RLock()
	return art.cmdList
}
func (art *Actor) SetCmdList(cmd uint32) {
	defer art.cmdLock.Unlock()
	art.cmdLock.Lock()
	if art.cmdList == nil {
		art.cmdList = make([]uint32, 0)
	}
	art.cmdList = append(art.cmdList, cmd)
}
