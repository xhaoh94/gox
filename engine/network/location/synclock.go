package location

import (
	"errors"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/commonhelper"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	SyncLocation struct {
	}
)

func (lock *SyncLocation) get(datas []uint32, excludeIDs []uint) []LocationData {
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
func (lock *SyncLocation) call(entity types.IServiceEntity, cmd uint32, msg any, response any) error {
	session := gox.NetWork.GetSessionByAddr(entity.GetInteriorAddr())
	if session == nil {
		return errors.New("session is nil")
	}
	return session.CallByCmd(cmd, msg, response)
}
