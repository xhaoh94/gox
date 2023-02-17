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
		syncWg sync.WaitGroup
	}
)

func (lock *SyncLocation) Lock() {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.Eid))
	lock.syncWg.Add(len(entitys))
	for _, v := range entitys {
		go lock.syncCall(5, v, consts.LocationLock, &LocationLock{Lock: true}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) UnLock() {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.Eid))
	lock.syncWg.Add(len(entitys))
	for _, v := range entitys {
		go lock.syncCall(5, v, consts.LocationLock, &LocationLock{Lock: false}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Add(datas []LocationData) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.Eid))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		go lock.syncCall(5, entity, consts.LocationAdd, &LocationAdd{Datas: datas}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Remove(datas []uint32) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.Eid))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		go lock.syncCall(5, entity, consts.LocationAdd, &LocationRemove{Datas: datas}, &LocationReslut{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Get() *LocationAdd {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.Eid))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		response := &LocationAdd{}
		if lock.syncCall(2, entity, consts.LocationGet, nil, response) {
			return response
		}
	}
	lock.syncWg.Wait()
	return nil
}

func (lock *SyncLocation) syncCall(tryCnt int, entity types.IServiceEntity, cmd uint32, msg any, response any) bool {
	defer lock.syncWg.Done()
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
		if sendCnt >= tryCnt { //最多尝试请求5次
			break
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒重新请求
	}
	return false
}
