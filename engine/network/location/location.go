package location

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/codec"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
)

const (
	waitTime time.Duration = 200
)

type (
	LocationSystem struct {
		SyncLocation

		// lockWg      sync.WaitGroup
		lockOther sync.RWMutex
		//缓存非本服务器的实体对应的服务器ID（这里的数据是有可能存在偏差的，例如某个实体已经转移到其他服务器上面去，这里记录的还是上一个服务器的ID）
		otherLocationMap map[uint32]uint

		lockSelf sync.RWMutex
		//注册在本服务器的实体
		slefLocationMap map[uint32]uint
	}
)

func New() *LocationSystem {
	locationSystem := &LocationSystem{}
	gox.Location = locationSystem
	return locationSystem
}
func (location *LocationSystem) Init() {
	location.otherLocationMap = make(map[uint32]uint, 0)
	location.slefLocationMap = make(map[uint32]uint, 0)
	if gox.Config.Location {
		protoreg.BindCodec(LocationRelay, codec.MsgPack)
		protoreg.BindCodec(LocationGet, codec.MsgPack)
		protoreg.BindCodec(LocationRegister, codec.MsgPack)
		protoreg.RegisterRpcCmd(LocationRelay, location.RelayHandler)
		protoreg.RegisterRpcCmd(LocationGet, location.GetHandler)
		protoreg.Register(LocationRegister, location.RegisterHandler)
	}
}
func (location *LocationSystem) Start() {
}

func (location *LocationSystem) Stop() {
}

func (location *LocationSystem) RelayHandler(ctx context.Context, session types.ISession, req *LocationRelayRequire) (*LocationRelayResponse, error) {

	location.lockSelf.RLock()
	_, ok := location.slefLocationMap[req.LocationID]
	location.lockSelf.RUnlock()
	relayResponse := &LocationRelayResponse{IsSuc: ok}
	if ok { //如果此实体注册在本服务器，那么进行逻辑处理，不然直接返回
		cmd := req.CMD
		require := protoreg.GetRequireByCmd(cmd)
		if err := session.Codec(cmd).Unmarshal(req.Require, require); err != nil {
			return relayResponse, nil
		}
		response, err := protoreg.Call(cmd, ctx, session, require)
		if err != nil || !req.IsCall {
			return relayResponse, nil
		}
		if msgData, err := session.Codec(cmd).Marshal(response); err == nil {
			relayResponse.Response = msgData
		}
	}
	return relayResponse, nil
}
func (location *LocationSystem) GetHandler(ctx context.Context, session types.ISession, req *LocationGetRequire) (*LocationGetResponse, error) {
	datas := make([]LocationData, 0)
	if len(location.slefLocationMap) > 0 && len(req.IDs) > 0 {
		defer location.lockSelf.RUnlock()
		location.lockSelf.RLock()
		for _, k := range req.IDs {
			if v, ok := location.slefLocationMap[k]; ok {
				datas = append(datas, LocationData{LocationID: k, AppID: v})
			}
		}
	}
	return &LocationGetResponse{Datas: datas}, nil
}
func (location *LocationSystem) RegisterHandler(ctx context.Context, session types.ISession, req *LocationRegisterRequire) {
	if req.AppID > 0 && req.AppID != gox.Config.AppID && len(req.LocationIDs) > 0 {
		location.lockOther.Lock()
		for _, locationID := range req.LocationIDs {
			if req.IsRegister {
				location.otherLocationMap[locationID] = req.AppID
			} else {
				delete(location.otherLocationMap, locationID)
			}
		}
		location.lockOther.Unlock()
	}
}

func (location *LocationSystem) add(Datas []LocationData) {
	if len(Datas) > 0 {
		location.lockOther.Lock()
		for _, v := range Datas {
			location.otherLocationMap[v.LocationID] = v.AppID
		}
		location.lockOther.Unlock()
	}
}
func (location *LocationSystem) del(Datas []uint32) {
	if len(Datas) > 0 && len(location.otherLocationMap) > 0 {
		location.lockOther.Lock()
		for _, v := range Datas {
			delete(location.otherLocationMap, v)
		}
		location.lockOther.Unlock()
	}
}

// 更新location所在的服务器，
// excludeIDs 排除的服务器列表
func (location *LocationSystem) updateLocationToAppID(locationID uint32, excludeServiceIDs []uint) {
	datas := location.SyncLocation.get([]uint32{locationID}, excludeServiceIDs)
	location.add(datas)
}

func (location *LocationSystem) Register(entity types.ILocation) {
	if !gox.Config.Location {
		logger.Error().Msg("没有启动Location的服务器不可以添加实体")
		return
	}
	locationID := entity.LocationID()
	if locationID == 0 {
		logger.Error().Msg("Location没有初始化ID")
		return
	}
	go entity.Init(entity)

	location.lockSelf.Lock()
	location.slefLocationMap[locationID] = gox.Config.AppID
	logger.Debug().Uint32("LocationID", locationID).Uint("AppID", gox.Config.AppID).Msg("注册Location")
	location.lockSelf.Unlock()

	location.SyncLocation.register(true, []uint32{locationID})
}
func (location *LocationSystem) Registers(entitys []types.ILocation) {
	if !gox.Config.Location {
		logger.Error().Msg("没有启动Location的服务器不可以添加实体")
		return
	}
	datas := make([]uint32, 0)
	location.lockSelf.Lock()
	for _, entity := range entitys {
		locationID := entity.LocationID()
		if locationID == 0 {
			logger.Error().Msg("Location没有初始化ID")
			continue
		}
		go entity.Init(entity)
		location.slefLocationMap[locationID] = gox.Config.AppID
		logger.Debug().Uint32("LocationID", locationID).Uint("AppID", gox.Config.AppID).Msg("注册Location")
		datas = append(datas, locationID)
	}
	location.lockSelf.Unlock()

	location.SyncLocation.register(true, datas)
}
func (location *LocationSystem) UnRegister(entity types.ILocation) {
	if !gox.Config.Location {
		logger.Error().Msg("没有启动Location的服务器不可以删除实体")
		return
	}
	if len(location.slefLocationMap) == 0 {
		return
	}
	locationID := entity.LocationID()
	if locationID == 0 {
		logger.Error().Msg("Location没有初始化ID")
		return
	}
	location.lockSelf.Lock()
	delete(location.slefLocationMap, locationID)
	logger.Debug().Uint32("LocationID", locationID).Msg("移除Location")
	location.lockSelf.Unlock()
	location.SyncLocation.register(false, []uint32{locationID})
	go entity.Destroy(entity)
}
func (location *LocationSystem) UnRegisters(entitys []types.ILocation) {
	if !gox.Config.Location {
		logger.Error().Msg("没有启动Location的服务器不可以删除实体")
		return
	}
	if len(location.otherLocationMap) == 0 {
		return
	}
	datas := make([]uint32, 0)
	location.lockSelf.Lock()
	for _, entity := range entitys {
		locationID := entity.LocationID()
		if locationID == 0 {
			logger.Error().Msg("Location没有初始化ID")
			continue
		}
		delete(location.slefLocationMap, locationID)
		logger.Debug().Uint32("LocationID", locationID).Msg("移除Location")
		datas = append(datas, locationID)
	}
	location.lockSelf.Unlock()

	location.SyncLocation.register(false, datas)

	for _, entity := range entitys {
		go entity.Destroy(entity)
	}
}
func (location *LocationSystem) ServiceClose(appID uint) {
	if appID == gox.Config.AppID {
		clear(location.slefLocationMap)
		return
	}
	if len(location.otherLocationMap) == 0 {
		return
	}
	defer location.lockOther.Unlock()
	location.lockOther.Lock()
	for k, v := range location.otherLocationMap {
		if v == appID {
			delete(location.otherLocationMap, k)
		}
	}
}

func (location *LocationSystem) Send(locationID uint32, require any) {
	if locationID == 0 {
		logger.Error().Msg("LocationSend LocationID不能为空")
		return
	}

	go func(_locationID uint32, _require any) {
		loopCnt := 0
		cmd := cmdhelper.ToCmd(_require, nil, _locationID)
		excludeIDs := make([]uint, 0)
		waitFn := func(id uint) {
			location.del([]uint32{_locationID})
			excludeIDs = append(excludeIDs, id)
			time.Sleep(time.Millisecond * waitTime) //等待0.2秒
		}
		for {
			loopCnt++
			if loopCnt > 3 {
				logger.Error().Msg("LocationSend:超出尝试发送上限")
				return
			}
			location.lockSelf.RLock()
			_, ok := location.slefLocationMap[_locationID]
			location.lockSelf.RUnlock()

			if ok {
				if !protoreg.HasBindCallBack(cmd) { //可能实体转移到其他服务器了，等待一下，再重新请求
					logger.Warn().Uint32("CMD", cmd).Msg("LocationSend 发送消息,找不到对应的CMD处理方法")
					continue
				}
				session := gox.NetWork.GetSessionByAppID(gox.Config.AppID)
				_, err := protoreg.Call(cmd, gox.Ctx, session, _require)
				if err != nil {
					logger.Warn().Err(err).Uint32("CMD", cmd).Msg("LocationSend 发送消息失败")
				}
				return
			}

			location.lockOther.RLock()
			id, ok := location.otherLocationMap[_locationID]
			location.lockOther.RUnlock()
			if !ok {
				location.updateLocationToAppID(_locationID, excludeIDs)
				continue
			}
			session := gox.NetWork.GetSessionByAppID(id)
			if session == nil {
				waitFn(id)
				continue
			}

			msgData, err := session.Codec(cmd).Marshal(_require)
			if err != nil {
				logger.Warn().Err(err).Uint32("CMD", cmd).Msg("LocationSend 序列化失败")
				return
			}
			tmpResponse := location.relay(session, cmd, _locationID, false, msgData)
			if tmpResponse == nil {
				return
			}
			if !tmpResponse.IsSuc { //可能实体转移到其他服务器了，等待一下，再重新请求
				waitFn(id)
				continue
			}
			return
		}
	}(locationID, require)
}
func (location *LocationSystem) Call(locationID uint32, require any, response any) error {
	if locationID == 0 {
		return errors.New("LocationCall LocationID不能为空")
	}
	excludeIDs := make([]uint, 0)
	waitFn := func(id uint) {
		location.del([]uint32{locationID})
		excludeIDs = append(excludeIDs, id)
		time.Sleep(time.Millisecond * waitTime) //等待一下
	}

	loopCnt := 0
	cmd := cmdhelper.ToCmd(require, response, locationID)
	for {
		loopCnt++
		if loopCnt > 3 {
			return errors.New("LocationCall:超出尝试发送上限")
		}

		location.lockSelf.RLock()
		_, ok := location.slefLocationMap[locationID]
		location.lockSelf.RUnlock()

		if ok {
			if !protoreg.HasBindCallBack(cmd) {
				logger.Warn().Uint32("CMD", cmd).Msg("LocationCall 发送消息,找不到对应的CMD处理方法")
				continue
			}
			session := gox.NetWork.GetSessionByAppID(gox.Config.AppID)
			resp, err := protoreg.Call(cmd, gox.Ctx, session, require)
			if err != nil {
				return err
			}
			if resp != nil {
				commonhelper.ReplaceValue(response, resp)
			}
			return nil
		}

		location.lockOther.RLock()
		id, ok := location.otherLocationMap[locationID]
		location.lockOther.RUnlock()
		if !ok {
			location.updateLocationToAppID(locationID, excludeIDs)
			continue
		}
		session := gox.NetWork.GetSessionByAppID(id)
		if session == nil {
			waitFn(id)
			continue
		}

		msgData, err := session.Codec(cmd).Marshal(require)
		if err != nil {
			return err
		}

		tmpResponse := location.relay(session, cmd, locationID, true, msgData)
		if tmpResponse == nil {
			return errors.New("转发消息失败")
		}
		if err != nil {
			return err
		}
		if !tmpResponse.IsSuc { //可能实体转移到其他服务器了，等待一下，再重新请求
			waitFn(id)
			continue
		}
		if len(tmpResponse.Response) > 0 {
			if err := session.Codec(cmd).Unmarshal(tmpResponse.Response, response); err != nil {
				return err
			}
		}
		return nil
	}
}
func (location *LocationSystem) Broadcast(locationIDs []uint32, require any) {
	for _, locationID := range locationIDs {
		location.Send(locationID, require)
	}
}
