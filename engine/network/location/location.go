package location

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	LocationSystem struct {
		SyncLocation

		// lockWg      sync.WaitGroup
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
	// protoreg.RegisterRpcCmd(consts.LocationLock, location.LockHandler)
	protoreg.RegisterRpcCmd(consts.LocationForward, location.ForwardHandler)
	protoreg.RegisterRpcCmd(consts.LocationGet, location.GetHandler)
	// protoreg.RegisterRpcCmd(consts.LocationAdd, location.AddHandler)
	// protoreg.RegisterRpcCmd(consts.LocationRemove, location.RemoveHandler)

}
func (location *LocationSystem) Start() {
}

func (location *LocationSystem) Stop() {
}

func (location *LocationSystem) ForwardHandler(ctx context.Context, session types.ISession, req *LocationForwardRequire) (*LocationForwardResponse, error) {
	forwardResponse := &LocationForwardResponse{}
	cmd := req.CMD
	forwardResponse.IsSuc = gox.Event.HasBind(cmd)
	if forwardResponse.IsSuc {
		require := protoreg.GetRequireByCmd(cmd)
		if err := session.Codec().Unmarshal(req.Require, require); err != nil {
			return forwardResponse, nil
		}
		response, err := cmdhelper.CallEvt(cmd, ctx, session, require)
		if err != nil || !req.IsCall {
			return forwardResponse, nil
		}
		if msgData, err := session.Codec().Marshal(response); err == nil {
			forwardResponse.Response = msgData
		}
	}
	return forwardResponse, nil
}
func (location *LocationSystem) GetHandler(ctx context.Context, session types.ISession, req *LocationGetRequire) (*LocationGetResponse, error) {
	datas := make([]LocationData, 0)
	if len(location.locationMap) > 0 && len(req.IDs) > 0 {
		defer location.lock.RUnlock()
		location.lock.RLock()
		for _, k := range req.IDs {
			if v, ok := location.locationMap[k]; ok {
				datas = append(datas, LocationData{LocationID: k, AppID: v})
			}
		}
	}
	return &LocationGetResponse{Datas: datas}, nil
}

//	func (location *LocationSystem) AddHandler(ctx context.Context, session types.ISession, req *LocationAddRequire) (*LocationAddResponse, error) {
//		if req != nil && req.Datas != nil {
//			location.add(req.Datas)
//		}
//		return &LocationAddResponse{}, nil
//	}
//
//	func (location *LocationSystem) RemoveHandler(ctx context.Context, session types.ISession, req *LocationRemoveRequire) (*LocationRemoveResponse, error) {
//		if req != nil && req.IDs != nil {
//			location.del(req.IDs)
//		}
//		return &LocationRemoveResponse{}, nil
//	}

func (location *LocationSystem) add(Datas []LocationData, isLock bool) {
	if len(Datas) > 0 {
		if isLock {
			defer location.lock.Unlock()
			location.lock.Lock()
		}
		for _, v := range Datas {
			xlog.Debug("新增LocationID:%d,AppID:%d", v.LocationID, v.AppID)
			location.locationMap[v.LocationID] = v.AppID
		}
	}
}
func (location *LocationSystem) del(Datas []uint32, isLock bool) {
	if len(Datas) > 0 && len(location.locationMap) > 0 {
		if isLock {
			defer location.lock.Unlock()
			location.lock.Lock()
		}
		for _, v := range Datas {
			if _, ok := location.locationMap[v]; ok {
				delete(location.locationMap, v)
				xlog.Debug("删除LocationID:%d", v)
			}
		}
	}
}

func (location *LocationSystem) UpdateLocationToAppID(locationID uint32, excludeIDs []uint) {
	defer location.lock.Unlock()
	location.lock.Lock()
	_, ok := location.locationMap[locationID]
	if ok {
		return
	}
	datas := location.SyncLocation.Get([]uint32{locationID}, excludeIDs)
	location.add(datas, false)
}
func (location *LocationSystem) UpdateLocationToAppIDs(locationIDs []uint32, excludeIDs []uint) {

	location.lock.Lock()
	defer location.lock.Unlock()

	reqIDs := make([]uint32, 0)
	for _, locationID := range locationIDs {
		if _, ok := location.locationMap[locationID]; !ok {
			reqIDs = append(reqIDs, locationID)
		}
	}
	if len(reqIDs) == 0 {
		return
	}

	datas := location.SyncLocation.Get(reqIDs, excludeIDs)
	location.add(datas, false)
}

func (location *LocationSystem) Add(entity types.ILocationEntity) {
	if !gox.AppConf.Location {
		xlog.Error("没有启动Location的服务器不可以添加实体")
		return
	}
	aid := entity.LocationID()
	if aid == 0 {
		xlog.Error("Location没有初始化ID")
		return
	}
	go entity.Init(entity)
	datas := []LocationData{{LocationID: aid, AppID: gox.AppConf.AppID}}
	location.add(datas, true)
	// location.SyncLocation.Add(datas)
}
func (location *LocationSystem) Adds(entitys []types.ILocationEntity) {
	if !gox.AppConf.Location {
		xlog.Error("没有启动Location的服务器不可以添加实体")
		return
	}
	datas := make([]LocationData, 0)
	for _, entity := range entitys {
		aid := entity.LocationID()
		if aid == 0 {
			xlog.Error("Location没有初始化ID")
			return
		}
		go entity.Init(entity)
		datas = append(datas, LocationData{LocationID: aid, AppID: gox.AppConf.AppID})
	}
	if len(datas) == 0 {
		return
	}
	location.add(datas, true)
	// location.SyncLocation.Add(datas)
}
func (location *LocationSystem) Del(entity types.ILocationEntity) {
	if !gox.AppConf.Location {
		xlog.Error("没有启动Location的服务器不可以删除实体")
		return
	}
	if len(location.locationMap) == 0 {
		return
	}
	aid := entity.LocationID()
	if aid == 0 {
		xlog.Error("Location没有初始化ID")
		return
	}
	datas := []uint32{aid}
	location.del(datas, true)
	// location.SyncLocation.Remove(datas)
	go entity.Destroy(entity)
}
func (location *LocationSystem) Dels(entitys []types.ILocationEntity) {
	if !gox.AppConf.Location {
		xlog.Error("没有启动Location的服务器不可以删除实体")
		return
	}
	if len(location.locationMap) == 0 {
		return
	}
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
	location.del(datas, true)
	// location.SyncLocation.Remove(datas)
	for _, entity := range entitys {
		go entity.Destroy(entity)
	}
}
func (location *LocationSystem) ServiceClose(appID uint) {
	if len(location.locationMap) == 0 {
		return
	}
	defer location.lock.Unlock()
	location.lock.Lock()
	for k, v := range location.locationMap {
		if v == appID {
			xlog.Debug("删除Location:%d", k)
			delete(location.locationMap, k)
		}
	}
}

func (location *LocationSystem) Send(locationID uint32, require any) {
	if locationID == 0 {
		xlog.Error("LocationSend 传入locationID不能为空")
		return
	}

	go func() {
		loopCnt := 0
		cmd := cmdhelper.ToCmd(require, nil, locationID)
		excludeIDs := make([]uint, 0)
		waitFn := func(id uint) {
			location.del([]uint32{locationID}, true)
			excludeIDs = append(excludeIDs, id)
			time.Sleep(time.Millisecond * 300) //等待0.3秒
		}
		for {
			loopCnt++
			if loopCnt > 3 {
				return
			}
			location.lock.RLock()
			id, ok := location.locationMap[locationID]
			location.lock.RUnlock()
			if !ok {
				location.UpdateLocationToAppID(locationID, excludeIDs)
				continue
			}
			session := gox.NetWork.GetSessionByAppID(id)
			if session == nil {
				waitFn(id)
				continue
			}
			if id == gox.AppConf.AppID {
				if !gox.Event.HasBind(cmd) {
					waitFn(id)
					continue
				}
				_, err := cmdhelper.CallEvt(cmd, gox.Ctx, session, require)
				if err != nil {
					return
				}
				xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
				return
			}

			msgData, err := session.Codec().Marshal(require)
			if err != nil {
				return
			}
			tmpRequire := &LocationForwardRequire{}
			tmpRequire.CMD = cmd
			tmpRequire.IsCall = false
			tmpRequire.Require = msgData
			tmpResponse := &LocationForwardResponse{}
			err = session.CallByCmd(consts.LocationForward, tmpRequire, tmpResponse)
			if err != nil {
				return
			}

			if !tmpResponse.IsSuc {
				waitFn(id)
				continue
			}
			return
		}
	}()
}
func (location *LocationSystem) Call(locationID uint32, require any, response any) error {
	if locationID == 0 {
		xlog.Error("LocationCall传入locationID不能为空")
		return errors.New("LocationCall:传入locationID不能为空")
	}
	excludeIDs := make([]uint, 0)
	waitFn := func(id uint) {
		location.del([]uint32{locationID}, true)
		excludeIDs = append(excludeIDs, id)
		time.Sleep(time.Millisecond * 300) //等待0.3秒
	}

	loopCnt := 0
	cmd := cmdhelper.ToCmd(require, response, locationID)
	for {
		loopCnt++
		if loopCnt > 3 {
			return errors.New("LocationCall:超出尝试发送上限")
		}
		location.lock.RLock()
		id, ok := location.locationMap[locationID]
		location.lock.RUnlock()
		if !ok {
			location.UpdateLocationToAppID(locationID, excludeIDs)
			continue
		}
		session := gox.NetWork.GetSessionByAppID(id)
		if session == nil {
			waitFn(id)
			continue
		}

		if id == gox.AppConf.AppID {
			if !gox.Event.HasBind(cmd) {
				waitFn(id)
				continue
			}
			resp, err := cmdhelper.CallEvt(cmd, gox.Ctx, session, require)
			if err != nil {
				return err
			}
			if resp != nil {
				commonhelper.ReplaceValue(response, resp)
			}
			return nil
		}

		msgData, err := session.Codec().Marshal(require)
		if err != nil {
			return err
		}
		tmpRequire := &LocationForwardRequire{}
		tmpRequire.CMD = cmd
		tmpRequire.IsCall = true
		tmpRequire.Require = msgData
		tmpResponse := &LocationForwardResponse{}
		err = session.CallByCmd(consts.LocationForward, tmpRequire, tmpResponse)

		if err != nil {
			return err
		}
		if !tmpResponse.IsSuc {
			waitFn(id)
			continue
		}
		if len(tmpResponse.Response) > 0 {
			if err := session.Codec().Unmarshal(tmpResponse.Response, response); err != nil {
				return err
			}
		}
		return nil
	}
}
func (as *LocationSystem) Broadcast(locationIDs []uint32, require any) {
	for _, locationID := range locationIDs {
		as.Send(locationID, require)
	}
}
