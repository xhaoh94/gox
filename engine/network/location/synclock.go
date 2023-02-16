package location

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	SyncLocation struct {
		wg sync.WaitGroup
	}
)

//	func (lock *SyncLock) Wait() {
//		lock.wg.Wait()
//	}
func (lock *SyncLocation) Lock() {
	entitys := gox.NetWork.GetServiceEntitys()
	lock.wg.Add(len(entitys))
	for _, v := range entitys {
		go lock.syncSend(v, consts.LocationLock, &LocationLock{Lock: true})
	}
	lock.wg.Wait()
}
func (lock *SyncLocation) UnLock() {
	entitys := gox.NetWork.GetServiceEntitys()

	for _, v := range entitys {
		if v.GetID() == gox.AppConf.Eid {
			continue
		}
		lock.wg.Add(1)
		go lock.syncSend(v, consts.LocationLock, &LocationLock{Lock: false})
	}
	lock.wg.Wait()
}
func (lock *SyncLocation) Add(actorID uint32, appID uint) {
	entitys := gox.NetWork.GetServiceEntitys()
	for _, v := range entitys {
		if v.GetID() == gox.AppConf.Eid {
			continue
		}
		lock.wg.Add(1)
		go lock.syncSend(v, consts.LocationRefresh, &LocationEntity{ActorID: actorID, AppID: appID})
	}
	lock.wg.Wait()
}
func (lock *SyncLocation) Remove(actorID uint32) {
	entitys := gox.NetWork.GetServiceEntitys()
	for _, v := range entitys {
		if v.GetID() == gox.AppConf.Eid {
			continue
		}
		lock.wg.Add(1)
		go lock.syncSend(v, consts.LocationRefresh, &LocationEntity{ActorID: actorID, AppID: 0})
	}
	lock.wg.Wait()
}
func (lock *SyncLocation) GetEntitys() *LocationInit {
	entitys := gox.NetWork.GetServiceEntitys()

	for _, v := range entitys {
		if v.GetID() == gox.AppConf.Eid {
			continue
		}
		var response *LocationInit
		if lock.syncCall(v, consts.LocationInit, nil, response) {
			return response
		}
	}
	return nil
}

func (lock *SyncLocation) syncSend(entity types.IServiceEntity, cmd uint32, msg any) {
	defer lock.wg.Done()
	session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return
	}
	sendCnt := 0
	for {
		sendCnt++
		if session.CallByCmd(cmd, msg, nil).Await() {
			break
		}
		if sendCnt >= 5 { //最多尝试请求5次
			break
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒重新请求
	}
}
func (lock *SyncLocation) syncCall(entity types.IServiceEntity, cmd uint32, msg any, response any) bool {
	session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return false
	}
	sendCnt := 0
	for {
		sendCnt++
		if session.CallByCmd(cmd, msg, response).Await() {
			return true
		}
		if sendCnt >= 5 { //最多尝试请求5次
			return false
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒重新请求
	}
	return false
}
