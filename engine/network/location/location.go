package location

import (
	"context"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	LocationSystem struct {
		gox.Module
		SyncLocation

		lockWg      sync.WaitGroup
		lock        sync.RWMutex
		locationMap map[uint32]uint
	}
)

func New() *LocationSystem {
	locationSystem := &LocationSystem{}
	gox.Location = locationSystem
	return locationSystem
}
func (location *LocationSystem) Init() {
	location.locationMap = make(map[uint32]uint, 0)
	protoreg.RegisterRpcCmd(consts.LocationLock, location.LockHandler)
	protoreg.RegisterRpcCmd(consts.LocationGet, location.GetHandler)
	protoreg.RegisterRpcCmd(consts.LocationAdd, location.AddHandler)
	protoreg.RegisterRpcCmd(consts.LocationRemove, location.RemoveHandler)

}
func (location *LocationSystem) Start() {

}

func (location *LocationSystem) Stop() {

}

func (location *LocationSystem) LockHandler(ctx context.Context, req *LocationLockRequire) *LocationLockResponse {

	if req.Lock {
		location.lockWg.Add(1)
	} else {
		location.lockWg.Done()
	}
	return &LocationLockResponse{}
}

func (location *LocationSystem) GetHandler(ctx context.Context, req *LocationGetRequire) *LocationGetResponse {
	defer location.lock.RUnlock()
	location.lock.RLock()
	datas := make([]LocationData, 0)
	for _, k := range req.IDs {
		if v, ok := location.locationMap[k]; ok {
			datas = append(datas, LocationData{LocationID: k, AppID: v})
		}
	}
	return &LocationGetResponse{Datas: datas}
}

func (location *LocationSystem) AddHandler(ctx context.Context, req *LocationAddRequire) *LocationAddResponse {
	if req != nil && req.Datas != nil {
		location.add(req.Datas)
	}
	return &LocationAddResponse{}
}
func (location *LocationSystem) RemoveHandler(ctx context.Context, req *LocationRemoveRequire) *LocationRemoveResponse {
	if req != nil && req.IDs != nil {
		location.del(req.IDs)
	}
	return &LocationRemoveResponse{}
}
func (location *LocationSystem) add(Datas []LocationData) {
	defer location.lock.Unlock()
	location.lock.Lock()
	for _, v := range Datas {
		xlog.Debug("新增LocationID:%d,AppID:%d", v.LocationID, v.AppID)
		location.locationMap[v.LocationID] = v.AppID
	}
}
func (location *LocationSystem) del(Datas []uint32) {
	defer location.lock.Unlock()
	location.lock.Lock()
	for _, v := range Datas {
		xlog.Debug("删除LocationID:%d", v)
		delete(location.locationMap, v)
	}
}

func (location *LocationSystem) GetAppID(locationID uint32) uint {
	location.lockWg.Wait()
	defer location.lock.RUnlock()
	location.lock.RLock()

	if id, ok := location.locationMap[locationID]; ok {
		return id
	}

	defer location.syncUnLock()
	location.syncLock()

	datas := location.SyncLocation.Get([]uint32{locationID})
	location.lock.Lock()
	var appID uint
	for _, v := range datas {
		location.locationMap[v.LocationID] = v.AppID
		if v.LocationID == locationID {
			appID = v.AppID
		}
	}
	location.lock.Unlock()
	return appID
}
func (location *LocationSystem) GetAppIDs(locationIDs []uint32) []uint {
	location.lockWg.Wait()
	defer location.lock.RUnlock()
	location.lock.RLock()
	AppIDs := make([]uint, 0)
	reqIDs := make([]uint32, 0)
	for _, locationID := range locationIDs {
		if id, ok := location.locationMap[locationID]; ok {
			AppIDs = append(AppIDs, id)
		} else {
			reqIDs = append(reqIDs, locationID)
		}
	}
	if len(reqIDs) == 0 {
		return AppIDs
	}

	defer location.syncUnLock()
	location.syncLock()

	datas := location.SyncLocation.Get(reqIDs)
	location.lock.Lock()
	for _, v := range datas {
		location.locationMap[v.LocationID] = v.AppID
		AppIDs = append(AppIDs, v.AppID)
	}
	location.lock.Unlock()
	return AppIDs
}

func (location *LocationSystem) Add(entity types.ILocationEntity) {
	aid := entity.LocationID()
	if aid == 0 {
		xlog.Error("Location没有初始化ID")
		return
	}
	entity.Init(entity)

	location.syncLock()
	datas := []LocationData{{LocationID: aid, AppID: gox.AppConf.AppID}}
	location.add(datas)
	location.SyncLocation.Add(datas)
	location.syncUnLock()
}
func (location *LocationSystem) Adds(entitys []types.ILocationEntity) {
	datas := make([]LocationData, 0)
	for _, entity := range entitys {
		aid := entity.LocationID()
		if aid == 0 {
			xlog.Error("Location没有初始化ID")
			return
		}
		entity.Init(entity)
		datas = append(datas, LocationData{LocationID: aid, AppID: gox.AppConf.AppID})
	}
	if len(datas) == 0 {
		return
	}
	location.syncLock()
	location.add(datas)
	location.SyncLocation.Add(datas)
	location.syncUnLock()
}
func (location *LocationSystem) Del(entity types.ILocationEntity) {
	aid := entity.LocationID()
	if aid == 0 {
		xlog.Error("Location没有初始化ID")
		return
	}
	location.syncLock()
	datas := []uint32{aid}
	location.del(datas)
	location.SyncLocation.Remove(datas)
	location.syncUnLock()
	entity.Destroy(entity)
}
func (location *LocationSystem) Dels(entitys []types.ILocationEntity) {
	datas := make([]uint32, 0)
	for _, entity := range entitys {
		aid := entity.LocationID()
		if aid == 0 {
			xlog.Error("Location没有初始化ID")
			return
		}
		datas = append(datas, aid)
	}
	if len(datas) == 0 {
		return
	}
	location.syncLock()
	location.del(datas)
	location.SyncLocation.Remove(datas)
	location.syncUnLock()
	for _, entity := range entitys {
		entity.Destroy(entity)
	}
}
func (location *LocationSystem) ServiceClose(appID uint) {
	defer location.lock.Unlock()
	location.lock.Lock()
	for k, v := range location.locationMap {
		if v == appID {
			xlog.Debug("删除Location:%d", k)
			delete(location.locationMap, k)
		}
	}
}

func (location *LocationSystem) syncLock() {
	location.lockWg.Add(1)
	location.SyncLocation.Lock()
}
func (location *LocationSystem) syncUnLock() {
	location.SyncLocation.UnLock()
	location.lockWg.Done()
}

func (as *LocationSystem) Send(locationID uint32, msg interface{}) bool {
	if locationID == 0 {
		xlog.Error("LocationSend 传入locationID不能为空")
		return false
	}
	loopCnt := 0
	cmd := cmdhelper.ToCmd(msg, nil, locationID)
	for {
		loopCnt++
		if loopCnt > 5 {
			return false
		}
		if id := as.GetAppID(locationID); id > 0 {
			if session := gox.NetWork.GetSessionByAppID(id); session != nil {
				if id == gox.AppConf.AppID {
					if _, err := cmdhelper.CallEvt(cmd, gox.Ctx, session, msg); err == nil {
						return true
					} else {
						xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
					}
				} else {
					return session.Send(cmd, msg)
				}
			}
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
}
func (as *LocationSystem) Call(locationID uint32, msg interface{}, response interface{}) types.IRpcx {
	if locationID == 0 {
		xlog.Error("LocationCall传入locationID不能为空")
		return rpc.NewEmptyRpcx()
	}

	loopCnt := 0
	cmd := cmdhelper.ToCmd(msg, response, locationID)
	for {
		loopCnt++
		if loopCnt > 5 {
			return rpc.NewEmptyRpcx()
		}
		if id := as.GetAppID(locationID); id > 0 {
			if id == gox.AppConf.AppID {
				if response, err := cmdhelper.CallEvt(cmd, gox.Ctx, msg); err == nil {
					rpcx := rpc.NewRpcx(gox.Ctx, response)
					defer rpcx.Run(true)
					return rpcx
				} else {
					xlog.Warn("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
				}
			} else {
				if session := gox.NetWork.GetSessionByAppID(id); session != nil {
					return session.CallByCmd(cmd, msg, response)
				}
			}
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
}
func (as *LocationSystem) Broadcast(locationIDs []uint32, msg interface{}) {
	for _, locationID := range locationIDs {
		go as.Send(locationID, msg)
	}
}
