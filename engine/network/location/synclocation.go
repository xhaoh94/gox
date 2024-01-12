package location

import (
	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	SyncLocation struct {
	}
)

func (sl *SyncLocation) relay(session types.ISession, cmd uint32, locationID uint32, isCall bool, msgDatas []byte) *LocationRelayResponse {
	tmpRequire := &LocationRelayRequire{}
	tmpRequire.LocationID = locationID
	tmpRequire.CMD = cmd
	tmpRequire.IsCall = isCall
	tmpRequire.Require = msgDatas
	tmpResponse := &LocationRelayResponse{}
	if err := session.CallByCmd(LocationRelay, tmpRequire, tmpResponse); err != nil {
		logger.Warn().Err(err).Uint32("CMD", cmd).Msg("Location relay error")
		return nil
	}
	return tmpResponse
}
func (sl *SyncLocation) get(datas []uint32, excludeIDs []uint) []LocationData {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.Config.AppID), types.WithLocation(), types.WithExcludeIDs(excludeIDs))
	Datas := make([]LocationData, 0)
	for _, entity := range entitys {
		if len(datas) == 0 {
			break
		}
		session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
		if session == nil {
			continue
		}
		response := &LocationGetResponse{}
		err := session.CallByCmd(LocationGet, &LocationGetRequire{IDs: datas}, response)
		if err == nil && response.Datas != nil && len(response.Datas) > 0 {
			for _, v := range response.Datas {
				datas = commonhelper.DeleteSlice(datas, v.LocationID)
				Datas = append(Datas, LocationData{LocationID: v.LocationID, AppID: v.AppID})
			}
		}
	}
	return Datas
}
func (sl *SyncLocation) register(isRegister bool, datas []uint32) {
	entitys := gox.NetWork.GetServiceEntitys(types.WithExcludeID(gox.Config.AppID), types.WithLocation())
	for _, entity := range entitys {
		if len(datas) == 0 {
			break
		}
		session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
		if session == nil {
			continue
		}
		session.Send(LocationRegister, &LocationRegisterRequire{IsRegister: isRegister, AppID: gox.Config.AppID, LocationIDs: datas})
	}
}
