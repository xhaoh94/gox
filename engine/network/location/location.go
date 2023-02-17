package location

import (
	"context"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	LocationSystem struct {
		gox.Module
		SyncLocation

		lockWg          sync.WaitGroup
		lock            sync.RWMutex
		locationDataMap map[uint32]uint

		cacellock sync.RWMutex
		cacelMap  map[uint32]uint
	}
	LocationAdd struct {
		Datas []LocationData
	}
	LocationRemove struct {
		Datas []uint32
	}
	LocationLock struct {
		Lock bool
	}
	LocationData struct {
		ActorID uint32
		AppID   uint
	}
	LocationReslut struct {
	}
)

func New() *LocationSystem {
	locationSystem := &LocationSystem{}
	gox.Location = locationSystem
	return locationSystem
}
func (location *LocationSystem) Init() {
	location.locationDataMap = make(map[uint32]uint, 0)
	location.cacelMap = make(map[uint32]uint, 0)
	location.RegisterRpc(consts.LocationLock, location.LockHandler)
	location.RegisterRpc(consts.LocationGet, location.GetHandler)
	location.RegisterRpc(consts.LocationAdd, location.AddHandler)
	location.RegisterRpc(consts.LocationRemove, location.RemoveHandler)

}
func (location *LocationSystem) Start() {
	location.SyncLocation.Lock()
	if locationData := location.SyncLocation.Get(); locationData != nil && len(locationData.Datas) > 0 {
		location.lock.Lock()
		for _, v := range locationData.Datas {
			xlog.Debug("新增Actor:%d,AppID:%d", v.ActorID, v.AppID)
			location.locationDataMap[v.ActorID] = v.AppID
		}
		defer location.lock.Unlock()
	}
	location.SyncLocation.UnLock()
}

func (location *LocationSystem) Stop() {

}

func (location *LocationSystem) LockHandler(ctx context.Context, req *LocationLock) *LocationReslut {

	if req.Lock {
		location.lockWg.Add(1)
	} else {
		location.lockWg.Done()
	}
	return &LocationReslut{}
}

func (location *LocationSystem) GetHandler(ctx context.Context) *LocationAdd {
	defer location.lock.RUnlock()
	location.lock.RLock()
	datas := make([]LocationData, len(location.locationDataMap))
	index := 0
	for aid, appId := range location.locationDataMap {
		datas[index] = LocationData{ActorID: aid, AppID: appId}
		index++
	}
	return &LocationAdd{Datas: datas}
}

func (location *LocationSystem) AddHandler(ctx context.Context, req *LocationAdd) *LocationReslut {
	if req != nil && req.Datas != nil {
		location.add(req.Datas)
	}
	return &LocationReslut{}
}
func (location *LocationSystem) RemoveHandler(ctx context.Context, req *LocationRemove) *LocationReslut {
	if req != nil && req.Datas != nil {
		location.del(req.Datas)
	}
	return &LocationReslut{}
}
func (location *LocationSystem) add(Datas []LocationData) {
	defer location.lock.Unlock()
	location.lock.Lock()
	for _, v := range Datas {
		xlog.Debug("新增Actor:%d,AppID:%d", v.ActorID, v.AppID)
		location.locationDataMap[v.ActorID] = v.AppID
	}
}
func (location *LocationSystem) del(Datas []uint32) {
	defer location.lock.Unlock()
	location.lock.Lock()
	for _, v := range Datas {
		xlog.Debug("删除Actor:%d", v)
		delete(location.locationDataMap, v)
	}
}

func (location *LocationSystem) RLockCacel(b bool) {
	if b {
		location.cacellock.RLock()
	} else {
		location.cacellock.RUnlock()
	}
}
func (location *LocationSystem) LockCacel(b bool) {
	if b {
		location.cacellock.Lock()
	} else {
		location.cacellock.Unlock()
	}
}
func (location *LocationSystem) GetAppID(actorID uint32) uint {
	if cacelId, ok := location.cacelMap[actorID]; ok {
		return cacelId
	} else {
		location.lockWg.Wait()
		defer location.lock.RUnlock()
		location.lock.RLock()
		if id, ok := location.locationDataMap[actorID]; ok && id != cacelId {
			location.cacellock.Lock()
			location.cacelMap[actorID] = id
			location.cacellock.Unlock()
			return id
		}
		return 0
	}
}

func (location *LocationSystem) Add(entity types.ILocationEntity) {
	aid := entity.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	if !entity.Init(entity) {
		return
	}

	location.syncLock()
	datas := []LocationData{{ActorID: aid, AppID: gox.AppConf.Eid}}
	location.add(datas)
	location.SyncLocation.Add(datas)
	location.syncUnLock()
}
func (location *LocationSystem) Adds(entitys []types.ILocationEntity) {
	datas := make([]LocationData, 0)
	for _, entity := range entitys {
		aid := entity.ActorID()
		if aid == 0 {
			xlog.Error("Actor没有初始化ID")
			return
		}
		if !entity.Init(entity) {
			return
		}
		datas = append(datas, LocationData{ActorID: aid, AppID: gox.AppConf.Eid})
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
	aid := entity.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	location.syncLock()
	datas := []uint32{aid}
	location.del(datas)
	location.SyncLocation.Remove(datas)
	location.syncUnLock()
	entity.Destroy()
}
func (location *LocationSystem) Dels(entitys []types.ILocationEntity) {
	datas := make([]uint32, 0)
	for _, entity := range entitys {
		aid := entity.ActorID()
		if aid == 0 {
			xlog.Error("Actor没有初始化ID")
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
		entity.Destroy()
	}
}
func (location *LocationSystem) ServiceClose(appID uint) {
	defer location.lock.Unlock()
	location.lock.Lock()
	for k, v := range location.locationDataMap {
		if v == appID {
			xlog.Debug("删除Actor:%d", k)
			delete(location.locationDataMap, k)
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
