package network

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
func (actor *Actor) AddActorFn(fn interface{}) {
	defer actor.fnLock.Unlock()
	actor.fnLock.Lock()
	if actor.fnList == nil {
		actor.fnList = make([]interface{}, 0)
	}
	actor.fnList = append(actor.fnList, fn)
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
