package gox

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	SyncLocation struct {
		syncWg sync.WaitGroup
	}
)

func (lock *SyncLocation) Lock() {
	entitys := NetWork.GetServiceEntitys()
	for _, v := range entitys {
		if v.GetID() == AppConf.Eid {
			continue
		}
		lock.syncWg.Add(1)
		go lock.syncCall(5, v, consts.LocationLock, &LocationLock{Lock: true}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) UnLock() {
	entitys := NetWork.GetServiceEntitys()

	for _, v := range entitys {
		if v.GetID() == AppConf.Eid {
			continue
		}
		lock.syncWg.Add(1)
		go lock.syncCall(5, v, consts.LocationLock, &LocationLock{Lock: false}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Add(actorID uint32, appID uint) {
	entitys := NetWork.GetServiceEntitys()
	for _, entity := range entitys {
		if entity.GetID() == AppConf.Eid {
			continue
		}
		lock.syncWg.Add(1)
		go lock.syncCall(5, entity, consts.LocationRefresh, &LocationData{ActorID: actorID, AppID: appID}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Remove(actorID uint32) {
	entitys := NetWork.GetServiceEntitys()
	for _, entity := range entitys {
		if entity.GetID() == AppConf.Eid {
			continue
		}
		lock.syncWg.Add(1)
		go lock.syncCall(5, entity, consts.LocationRefresh, &LocationData{ActorID: actorID, AppID: 0}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) GetDatas() *LocationAdd {
	entitys := NetWork.GetServiceEntitys()
	for _, entity := range entitys {
		if entity.GetID() == AppConf.Eid {
			continue
		}
		lock.syncWg.Add(1)
		response := &LocationAdd{}
		if lock.syncCall(2, entity, consts.LocationInit, nil, response) {
			return response
		}
	}
	lock.syncWg.Wait()
	return nil
}

func (lock *SyncLocation) syncCall(tryCnt int, entity types.IServiceEntity, cmd uint32, msg any, response any) bool {
	defer lock.syncWg.Done()
	session := NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return false
	}
	sendCnt := 0
	for {
		sendCnt++
		if session.CallByCmd(cmd, msg, response).Await() {
			return true
		}
		if sendCnt >= tryCnt { //最多尝试请求5次
			break
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒重新请求
	}
	return false
}
