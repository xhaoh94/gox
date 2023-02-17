package gox

import (
	"context"
	"sync"

	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	LocationSystem struct {
		Module
		SyncLocation
		ctx context.Context

		lockWg          sync.WaitGroup
		lock            sync.RWMutex
		locationDataMap map[uint32]uint

		cacellock sync.RWMutex
		cacelMap  map[uint32]uint
	}
	LocationAdd struct {
		Datas []LocationData
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

func New(ctx context.Context) *LocationSystem {
	return &LocationSystem{
		ctx: ctx,
	}
}
func (m *LocationSystem) Init() {
	m.locationDataMap = make(map[uint32]uint, 0)
	m.cacelMap = make(map[uint32]uint, 0)
	m.RegisterRpc(consts.LocationLock, m.LockHandler)
	m.RegisterRpc(consts.LocationRefresh, m.RefreshHandler)
	m.RegisterRpc(consts.LocationInit, m.InitHandler)
}
func (m *LocationSystem) Start() {
	m.SyncLocation.Lock()
	if locationData := m.SyncLocation.GetDatas(); locationData != nil && len(locationData.Datas) > 0 {
		m.lock.Lock()
		for _, v := range locationData.Datas {
			xlog.Debug("新增Actor:%d,AppID:%d", v.ActorID, v.AppID)
			m.locationDataMap[v.ActorID] = v.AppID
		}
		defer m.lock.Unlock()
	}
	m.SyncLocation.UnLock()
}

func (m *LocationSystem) Stop() {

}

func (m *LocationSystem) InitHandler(ctx context.Context) *LocationAdd {
	defer m.lock.RUnlock()
	m.lock.RLock()
	datas := make([]LocationData, len(m.locationDataMap))
	index := 0
	for aid, appId := range m.locationDataMap {
		datas[index] = LocationData{ActorID: aid, AppID: appId}
		index++
	}
	return &LocationAdd{Datas: datas}
}
func (m *LocationSystem) LockHandler(ctx context.Context, req *LocationLock) *LocationReslut {

	if req.Lock {
		xlog.Debug("Lock")
		m.lockWg.Add(1)
	} else {
		xlog.Debug("UnLock")
		m.lockWg.Done()
	}
	return &LocationReslut{}
}
func (m *LocationSystem) RefreshHandler(ctx context.Context, req *LocationData) *LocationReslut {
	xlog.Debug("RefreshHandler")
	m.addOrdel(req.ActorID, req.AppID)
	return &LocationReslut{}
}
func (m *LocationSystem) addOrdel(actorID uint32, appID uint) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if appID == 0 {
		xlog.Debug("删除Actor:%d", actorID)
		delete(m.locationDataMap, actorID)
	} else {
		xlog.Debug("新增Actor:%d,AppID:%d", actorID, appID)
		m.locationDataMap[actorID] = appID
	}
}

func (m *LocationSystem) RLockCacel(b bool) {
	if b {
		m.cacellock.RLock()
	} else {
		m.cacellock.RUnlock()
	}
}
func (m *LocationSystem) LockCacel(b bool) {
	if b {
		m.cacellock.Lock()
	} else {
		m.cacellock.Unlock()
	}
}

func (as *LocationSystem) GetAppId(actorID uint32) uint {
	if cacelId, ok := as.cacelMap[actorID]; ok {
		return cacelId
	} else {
		as.lockWg.Wait()
		defer as.lock.RUnlock()
		as.lock.RLock()
		if id, ok := as.locationDataMap[actorID]; ok && id != cacelId {
			as.cacellock.Lock()
			as.cacelMap[actorID] = id
			as.cacellock.Unlock()
			return id
		}
		return 0
	}
}

func (as *LocationSystem) Add(actor types.ILocationEntity) {
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	actor.Init()
	fnList := actor.GetFnList()
	if fnList == nil {
		xlog.Error("Actor没有注册回调函数")
		return
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := as.parseFn(aid, fn); cmd != 0 {
			actor.SetCmdList(cmd)
		}
	}

	as.syncLock()
	as.addOrdel(aid, AppConf.Eid)
	as.SyncLocation.Add(aid, AppConf.Eid)
	as.syncUnLock()
}
func (as *LocationSystem) Del(actor types.ILocationEntity) {
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	cmdList := actor.GetCmdList()
	for index := range cmdList {
		cmd := cmdList[index]
		protoreg.UnRegisterRType(cmd)
		Event.UnBind(cmd)
	}
	as.syncLock()
	as.addOrdel(aid, 0)
	as.SyncLocation.Remove(aid)
	as.syncUnLock()
	actor.Destroy()
}
func (as *LocationSystem) ServiceClose(appID uint) {
	as.lock.Lock()
	for k, v := range as.locationDataMap {
		if v == appID {
			xlog.Debug("删除Actor:%d", k)
			delete(as.locationDataMap, k)
		}
	}
	defer as.lock.Unlock()
}

func (m *LocationSystem) syncLock() {
	m.lockWg.Add(1)
	m.SyncLocation.Lock()
}
func (m *LocationSystem) syncUnLock() {
	m.SyncLocation.UnLock()
	m.lockWg.Done()
}
