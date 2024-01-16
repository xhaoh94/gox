package aoi

import (
	"github.com/xhaoh94/gox/engine/aoi/aoigrid"
	"github.com/xhaoh94/gox/engine/aoi/aoilink"
	"github.com/xhaoh94/gox/engine/types"
)

// 初始化一个九宫格AOI区域 (left right X轴长度)(top bottom Y轴长度) (gw gh 格子的宽高) (distance 搜索范围九宫格格数)
func NewAOIGridManager[T types.AOIKey](left, right, top, bottom float32, gw, gh, distance int) *aoigrid.AOIGridManager[T] {
	aoiMgr := &aoigrid.AOIGridManager[T]{
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

// 初始化一个十字链条AOI区域 distance距离
func NewAOILinkManager[T types.AOIKey](distance float32) *aoilink.AOILinkManager[T] {
	aoiMrg := &aoilink.AOILinkManager[T]{Distance: distance}
	aoiMrg.Init()
	return aoiMrg
}
