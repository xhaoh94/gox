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
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	LocationSystem struct {
		gox.Module
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
	protoreg.RegisterRpcCmd(consts.LocationAdd, location.AddHandler)
	protoreg.RegisterRpcCmd(consts.LocationRemove, location.RemoveHandler)

}
func (location *LocationSystem) Start() {

}

func (location *LocationSystem) Stop() {

}

// func (location *LocationSystem) LockHandler(ctx context.Context, req *LocationLockRequire) (*LocationLockResponse, error) {

//		if req.Lock {
//			location.lockWg.Add(1)
//		} else {
//			location.lockWg.Done()
//		}
//		return &LocationLockResponse{}, nil
//	}
func (location *LocationSystem) ForwardHandler(ctx context.Context, session types.ISession, req *LocationForwardRequire) (*LocationForwardResponse, error) {
	forwardResponse := &LocationForwardResponse{}
	cmd := req.CMD
	forwardResponse.IsSuc = gox.Event.HasBind(cmd)
	defer xlog.Debug("ForwardHandler %v", forwardResponse)
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

func (location *LocationSystem) AddHandler(ctx context.Context, session types.ISession, req *LocationAddRequire) (*LocationAddResponse, error) {
	if req != nil && req.Datas != nil {
		location.add(req.Datas)
	}
	return &LocationAddResponse{}, nil
}

func (location *LocationSystem) RemoveHandler(ctx context.Context, session types.ISession, req *LocationRemoveRequire) (*LocationRemoveResponse, error) {
	if req != nil && req.IDs != nil {
		location.del(req.IDs)
	}
	return &LocationRemoveResponse{}, nil
}
func (location *LocationSystem) add(Datas []LocationData) {
	if len(Datas) > 0 {
		defer location.lock.Unlock()
		location.lock.Lock()
		for _, v := range Datas {
			xlog.Debug("新增LocationID:%d,AppID:%d", v.LocationID, v.AppID)
			location.locationMap[v.LocationID] = v.AppID
		}
	}
}
func (location *LocationSystem) del(Datas []uint32) {
	if len(Datas) > 0 && len(location.locationMap) > 0 {
		defer location.lock.Unlock()
		location.lock.Lock()
		for _, v := range Datas {
			if _, ok := location.locationMap[v]; ok {
				delete(location.locationMap, v)
				xlog.Debug("删除LocationID:%d", v)
			}
		}
	}
}

func (location *LocationSystem) UpdateLocationToAppID(locationID uint32) {
	// defer location.syncUnLock()
	// location.syncLock()
	datas := location.SyncLocation.Get([]uint32{locationID})
	for _, v := range datas {
		location.locationMap[v.LocationID] = v.AppID
	}

}
func (location *LocationSystem) GetAppIDs(locationIDs []uint32) []uint {

	location.lock.Lock()
	defer location.lock.Unlock()

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

	// defer location.syncUnLock()
	// location.syncLock()

	datas := location.SyncLocation.Get(reqIDs)
	for _, v := range datas {
		location.locationMap[v.LocationID] = v.AppID
		AppIDs = append(AppIDs, v.AppID)
	}
	return AppIDs
}

func (location *LocationSystem) Add(entity types.ILocationEntity) {
	aid := entity.LocationID()
	if aid == 0 {
		xlog.Error("Location没有初始化ID")
		return
	}
	entity.Init(entity)
	// location.syncLock()
	datas := []LocationData{{LocationID: aid, AppID: gox.AppConf.AppID}}
	location.add(datas)
	// location.SyncLocation.Add(datas)
	// location.syncUnLock()
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
	// location.syncLock()
	location.add(datas)
	// location.SyncLocation.Add(datas)
	// location.syncUnLock()
}
func (location *LocationSystem) Del(entity types.ILocationEntity) {
	aid := entity.LocationID()
	if aid == 0 {
		xlog.Error("Location没有初始化ID")
		return
	}
	// location.syncLock()
	datas := []uint32{aid}
	location.del(datas)
	// location.SyncLocation.Remove(datas)
	// location.syncUnLock()
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
	// location.syncLock()
	location.del(datas)
	// location.SyncLocation.Remove(datas)
	// location.syncUnLock()
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

// func (location *LocationSystem) syncLock() {
// 	location.SyncLocation.Lock()
// }
// func (location *LocationSystem) syncUnLock() {
// 	location.SyncLocation.UnLock()
// }

func (location *LocationSystem) Send(locationID uint32, msg any) bool {
	if locationID == 0 {
		xlog.Error("LocationSend 传入locationID不能为空")
		return false
	}
	location.lock.Lock()
	defer location.lock.Unlock()
	loopCnt := 0
	cmd := cmdhelper.ToCmd(msg, nil, locationID)
	for {
		if id, ok := location.locationMap[locationID]; ok {
			if session := gox.NetWork.GetSessionByAppID(id); session != nil {
				if id == gox.AppConf.AppID {
					if gox.Event.HasBind(cmd) {
						if _, err := cmdhelper.CallEvt(cmd, gox.Ctx, session, msg); err == nil {
							return true
						} else {
							xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
							return false
						}
					} else {
						location.del([]uint32{locationID})
					}
				} else {
					msgData, err := session.Codec().Marshal(msg)
					if err != nil {
						return false
					}

					tmpRequire := &LocationForwardRequire{}
					tmpRequire.CMD = cmd
					tmpRequire.IsCall = false
					tmpRequire.Require = msgData
					tmpResponse := &LocationForwardResponse{}
					if err := session.CallByCmd(consts.LocationForward, tmpRequire, tmpResponse).Await(); err == nil {
						if tmpResponse.IsSuc {
							return true
						}
						location.del([]uint32{locationID})
					} else {
						return false
					}
				}
			}
		} else {
			location.UpdateLocationToAppID(locationID)
			continue
		}
		loopCnt++
		if loopCnt >= 3 {
			return false
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
}
func (location *LocationSystem) Call(locationID uint32, require any, response any) types.IRpcx {
	if locationID == 0 {
		xlog.Error("LocationCall传入locationID不能为空")
		return rpc.NewEmptyRpcx(errors.New("LocationCall:传入locationID不能为空"))
	}
	location.lock.Lock()
	defer location.lock.Unlock()

	loopCnt := 0
	cmd := cmdhelper.ToCmd(require, response, locationID)
	for {
		loopCnt++
		if loopCnt > 3 {
			return rpc.NewEmptyRpcx(errors.New("LocationCall:超出尝试发送上限"))
		}

		id, ok := location.locationMap[locationID]
		if !ok {
			location.UpdateLocationToAppID(locationID)
			continue
		}
		session := gox.NetWork.GetSessionByAppID(id)
		if session == nil {
			location.del([]uint32{locationID})
			time.Sleep(time.Millisecond * 500) //等待0.5秒
			continue
		}

		if id == gox.AppConf.AppID {
			if !gox.Event.HasBind(cmd) {
				location.del([]uint32{locationID})
				time.Sleep(time.Millisecond * 500) //等待0.5秒
				continue
			}
			resp, err := cmdhelper.CallEvt(cmd, gox.Ctx, session, require)
			if err != nil {
				return rpc.NewEmptyRpcx(err)
			}
			if resp != nil {
				commonhelper.ReplaceValue(response, resp)
			}
			rpcx := rpc.NewEmptyRpcx(nil)
			defer rpcx.Run(nil)
			return rpcx
		}

		msgData, err := session.Codec().Marshal(require)
		if err != nil {
			return rpc.NewEmptyRpcx(err)
		}
		tmpRequire := &LocationForwardRequire{}
		tmpRequire.CMD = cmd
		tmpRequire.IsCall = true
		tmpRequire.Require = msgData
		tmpResponse := &LocationForwardResponse{}
		xlog.Debug("LocationForwardRequire %v", tmpRequire)
		err = session.CallByCmd(consts.LocationForward, tmpRequire, tmpResponse).Await()
		xlog.Debug("LocationForwardResponse %v", tmpResponse)
		if err != nil {
			return rpc.NewEmptyRpcx(err)
		}
		if !tmpResponse.IsSuc {
			location.del([]uint32{locationID})
			time.Sleep(time.Millisecond * 500) //等待0.5秒
			continue
		}
		if len(tmpResponse.Response) > 0 {
			if err := session.Codec().Unmarshal(tmpResponse.Response, response); err != nil {
				return rpc.NewEmptyRpcx(err)
			}
		}
		rpcx := rpc.NewEmptyRpcx(nil)
		defer rpcx.Run(nil)
		return rpcx
	}
}
func (as *LocationSystem) Broadcast(locationIDs []uint32, msg interface{}) {
	for _, locationID := range locationIDs {
		go as.Send(locationID, msg)
	}
}
