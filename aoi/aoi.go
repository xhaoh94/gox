package aoi

import (
	"github.com/xhaoh94/gox/engine/aoi/aoibase"
	"github.com/xhaoh94/gox/engine/aoi/aoigrid"
	"github.com/xhaoh94/gox/engine/aoi/aoilink"
)

type IAOIManager interface {
	Enter(string, float32, float32)
	Leave(string)
	Update(string, float32, float32)
	Find(string) aoibase.IAOIResult
}

//初始化一个九宫格AOI区域 (left right X轴长度)(top bottom Y轴长度) (gw gh 格子的宽高) (distance 搜索范围九宫格格数)
func NewAOIGridManager(left, right, top, bottom, gw, gh, distance int) *aoigrid.AOIGridManager {
	aoiMgr := &aoigrid.AOIGridManager{
		Left:       left,
		Right:      right,
		Top:        top,
		Bottom:     bottom,
		GridWidth:  gw,
		GridHeight: gh,
		Distance:   distance,
	}
	aoiMgr.Init()
	return aoiMgr
}

//初始化一个十字链条AOI区域 distance距离
func NewAOILinkManager(distance float32) *aoilink.AOILinkManager {
	aoiMrg := &aoilink.AOILinkManager{Distance: distance}
	aoiMrg.Init()
	return aoiMrg
}
