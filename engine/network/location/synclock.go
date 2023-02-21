package location

import (
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
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
		go lock.syncCall(v, consts.LocationLock, &LocationLockRequire{Lock: true}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) UnLock() {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, v := range entitys {
		go lock.syncCall(v, consts.LocationLock, &LocationLockRequire{Lock: false}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Add(datas []LocationData) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		go lock.syncCall(entity, consts.LocationAdd, &LocationAddRequire{Datas: datas}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}
func (lock *SyncLocation) Remove(datas []uint32) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	lock.syncWg.Add(len(entitys))
	for _, entity := range entitys {
		go lock.syncCall(entity, consts.LocationAdd, &LocationRemoveRequire{IDs: datas}, &LocationLockResponse{})
	}
	lock.syncWg.Wait()
}

func (lock *SyncLocation) Get(datas []uint32) []LocationData {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.AppConf.AppID))
	Datas := make([]LocationData, 0)
	xlog.Debug("xxxxxxxxxxxxxxxxxxxxxxxGetAppID %v", datas)
	for _, entity := range entitys {
		if len(datas) == 0 {
			break
		}
		response := &LocationGetResponse{}
		xlog.Debug("xxxxxxxxxxxxxxxxxxxxxxxCall APPID %d, 开始 %v", entity.GetID(), datas)
		err := lock.call(entity, consts.LocationGet, &LocationGetRequire{IDs: datas}, response).Await()
		if err == nil && response.Datas != nil && len(response.Datas) > 0 {
			for _, v := range response.Datas {
				datas = commonhelper.DeleteSlice(datas, v.LocationID)
				Datas = append(Datas, LocationData{LocationID: v.LocationID, AppID: v.AppID})
			}
		}
		xlog.Debug("xxxxxxxxxxxxxxxxxxxxxxxCall APPID %d, 结束 %v", entity.GetID(), Datas)
	}
	xlog.Debug("xxxxxxxxxxxxxxxxxxxxxxxLocationData  %v", Datas)
	return Datas
}

func (lock *SyncLocation) syncCall(entity types.IServiceEntity, cmd uint32, msg any, response any) error {
	defer lock.syncWg.Done()
	session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return consts.Error_4
	}
	return session.CallByCmd(cmd, msg, response).Await()
}
func (lock *SyncLocation) call(entity types.IServiceEntity, cmd uint32, msg any, response any) types.IRpcx {
	session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return rpc.NewEmptyRpcx(consts.Error_4)
	}
	return session.CallByCmd(cmd, msg, response)
}
