package location

import (
	"errors"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	SyncLocation struct {
		// syncWg sync.WaitGroup
	}
)

// func (lock *SyncLocation) Lock() {
// 	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
// 	lock.syncWg.Add(len(entitys))
// 	for _, v := range entitys {
// 		go lock.syncCall(v, consts.LocationLock, &LocationLockRequire{Lock: true}, &LocationLockResponse{})
// 	}
// 	lock.syncWg.Wait()
// }
// func (lock *SyncLocation) UnLock() {
// 	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
// 	lock.syncWg.Add(len(entitys))
// 	for _, v := range entitys {
// 		go lock.syncCall(v, consts.LocationLock, &LocationLockRequire{Lock: false}, &LocationLockResponse{})
// 	}
// 	lock.syncWg.Wait()
// }

// func (lock *SyncLocation) Add(datas []LocationData) {
// 	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
// 	lock.syncWg.Add(len(entitys))
// 	for _, entity := range entitys {
// 		go lock.syncCall(entity, consts.LocationAdd, &LocationAddRequire{Datas: datas}, &LocationLockResponse{})
// 	}
// 	lock.syncWg.Wait()
// }
// func (lock *SyncLocation) Remove(datas []uint32) {
// 	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
// 	lock.syncWg.Add(len(entitys))
// 	for _, entity := range entitys {
// 		go lock.syncCall(entity, consts.LocationAdd, &LocationRemoveRequire{IDs: datas}, &LocationLockResponse{})
// 	}
// 	lock.syncWg.Wait()
// }

func (lock *SyncLocation) Get(datas []uint32, excludeIDs []uint) []LocationData {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.Config.AppID), types.WithLocation(), types.WithExcludeIDs(excludeIDs))
	Datas := make([]LocationData, 0)
	for _, entity := range entitys {
		if len(datas) == 0 {
			break
		}
		response := &LocationGetResponse{}
		err := lock.call(entity, LocationGet, &LocationGetRequire{IDs: datas}, response)
		if err == nil && response.Datas != nil && len(response.Datas) > 0 {
			for _, v := range response.Datas {
				datas = commonhelper.DeleteSlice(datas, v.LocationID)
				Datas = append(Datas, LocationData{LocationID: v.LocationID, AppID: v.AppID})
			}
		}
	}
	return Datas
}

//	func (lock *SyncLocation) syncCall(entity types.IServiceEntity, cmd uint32, msg any, response any) error {
//		defer lock.syncWg.Done()
//		session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
//		if session == nil {
//			return consts.Error_4
//		}
//		return session.CallByCmd(cmd, msg, response)
//	}
func (lock *SyncLocation) call(entity types.IServiceEntity, cmd uint32, msg any, response any) error {
	session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return errors.New("session is nil")
	}
	return session.CallByCmd(cmd, msg, response)
}
