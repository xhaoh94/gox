package location

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	SyncLocation struct {
		syncWg sync.WaitGroup
	}
)

func (lock *SyncLocation) Lock() {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, v := range entitys {
		go lock.syncCall(1, true, v, consts.LocationLock, &LocationLockRequire{Lock: true}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) UnLock() {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, v := range entitys {
		go lock.syncCall(1, true, v, consts.LocationLock, &LocationLockRequire{Lock: false}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Add(datas []LocationData) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		go lock.syncCall(1, true, entity, consts.LocationAdd, &LocationAddRequire{Datas: datas}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Remove(datas []uint32) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		go lock.syncCall(1, true, entity, consts.LocationAdd, &LocationRemoveRequire{IDs: datas}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}

func (lock *SyncLocation) Get(datas []uint32) []LocationData {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	Datas := make([]LocationData, 0)
	for _, entity := range entitys {
		if len(datas) == 0 {
			break
		}
		response := &LocationGetResponse{}
		if lock.syncCall(1, false, entity, consts.LocationGet, &LocationGetRequire{IDs: datas}, response) {
			if response != nil && response.Datas != nil {
				for _, v := range response.Datas {
					datas = commonhelper.DeleteSlice(datas, v.LocationID)
					Datas = append(Datas, LocationData{LocationID: v.LocationID, AppID: v.AppID})
				}
			}
		}
	}
	return Datas
}

func (lock *SyncLocation) syncCall(tryCnt int, isDone bool, entity types.IServiceEntity, cmd uint32, msg any, response any) bool {
	if isDone {
		defer lock.syncWg.Done()
	}
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
		if sendCnt >= tryCnt {
			break
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒重新请求
	}
	return false
}
