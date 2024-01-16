package aoigrid

import (
	"math"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	AOIGridManager[T types.AOIKey] struct {
		Left       float32 //区域左边界坐标
		Right      float32 //区域右边界坐标
		Top        float32 //区域上边界坐标
		Bottom     float32 //区域下边界坐标
		GridWidth  int     //格子的宽
		GridHeight int     //格子的高
		Distance   int     //搜索范围距离 ->当前周围 n格子范围

		cntsX        int                 //x方向格子的数量
		cntsY        int                 //y方向的格子数量
		grids        map[int]*AOIGrid[T] //当前区域中都有哪些格子，key=格子ID， value=格子对象
		unitToGridID map[T]int           //物体对应的格子
		lock         sync.RWMutex
	}
)

func (m *AOIGridManager[T]) Init() {
	m.grids = make(map[int]*AOIGrid[T], 0)
	m.unitToGridID = make(map[T]int, 0)
	m.cntsX = int(math.Ceil(float64(m.Right-m.Left) / float64(m.GridWidth)))
	m.cntsY = int(math.Ceil(float64(m.Bottom-m.Top) / float64(m.GridHeight)))
	//给AOI初始化区域中所有的格子
	for y := 0; y < m.cntsY; y++ {
		for x := 0; x < m.cntsX; x++ {
			//计算格子ID
			//格子编号：id = idy *nx + idx  (利用格子坐标得到格子编号)
			gid := y*m.cntsX + x
			//初始化一个格子放在AOI中的map里，key是当前格子的ID
			minx := m.Left + float32(x*m.GridWidth)
			maxx := m.Left + float32((x+1)*m.GridWidth)
			if maxx > m.Right {
				maxx = m.Right
			}
			miny := m.Top + float32(y*m.GridHeight)
			maxy := m.Left + float32((y+1)*m.GridHeight)
			if maxy > m.Bottom {
				maxy = m.Bottom
			}
			m.grids[gid] = newGrid[T](gid, minx, maxx, miny, maxy)
		}
	}
}
func (m *AOIGridManager[T]) Enter(unit T, x, y float32) {
	if x < m.Left || x > m.Right || y < m.Top || y > m.Bottom {
		logger.Warn().Str("Key", strhelper.ValToString(unit)).Float32("x", x).Float32("y", y).Msg("AOI_Enter超出边界")
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	if gID, ok := m.unitToGridID[unit]; ok {
		m.grids[gID].Remove(unit)
	}
	gID := m.getGridIDByPos(x, y)
	m.grids[gID].Add(unit)
	m.unitToGridID[unit] = gID
}
func (m *AOIGridManager[T]) Leave(unit T) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if gID, ok := m.unitToGridID[unit]; ok {
		m.grids[gID].Remove(unit)
		delete(m.unitToGridID, unit)
	}
}
func (m *AOIGridManager[T]) Update(unit T, x, y float32) {
	if x < m.Left || x > m.Right || y < m.Top || y > m.Bottom {
		logger.Warn().Str("Key", strhelper.ValToString(unit)).Float32("x", x).Float32("y", y).Msg("AOI_Update超出边界")
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	if oldID, ok := m.unitToGridID[unit]; ok {
		newID := m.getGridIDByPos(x, y)
		if oldID != newID {
			_, ok := m.grids[newID]
			if !ok {
				logger.Error().Int("GID", newID).Float32("x", x).Float32("y", y).Msg("AOI_Update超出边界")
				return
			}
			m.grids[oldID].Remove(unit)
			m.grids[newID].Add(unit)
			m.unitToGridID[unit] = newID
		}
	}
}

func (m *AOIGridManager[T]) Find(unit T) types.IAOIResult[T] {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ids := newResult[T]()
	if gID, ok := m.unitToGridID[unit]; ok {
		grids := m.getSurroundGridsByGid(gID)
		for i := range grids {
			ids.pushs(grids[i].GetIDs())
		}
	}
	return ids
}

// 根据格子的gID得到当前周边的九宫格信息
func (m *AOIGridManager[T]) getSurroundGridsByGid(gID int) (grids []*AOIGrid[T]) {
	//判断gID是否存在
	if _, ok := m.grids[gID]; !ok {
		return
	}
	idx := gID % m.cntsX
	idy := gID / m.cntsX
	minx := max(idx-m.Distance, 0)
	maxx := min(idx+m.Distance, m.cntsX-1)

	miny := max(idy-m.Distance, 0)
	maxy := min(idy+m.Distance, m.cntsY-1)

	for y := miny; y <= maxy; y++ {
		for x := minx; x <= maxx; x++ {
			id := gID + (x - idx) + ((y - idy) * m.cntsX)
			grids = append(grids, m.grids[id])
		}
	}
	return
}

// 通过横纵坐标获取对应的格子ID
func (m *AOIGridManager[T]) getGridIDByPos(x, y float32) int {
	gx := int(math.Ceil(float64(x-m.Left) / float64(m.GridWidth)))
	gy := int(math.Ceil(float64(y-m.Top) / float64(m.GridWidth)))
	return gy*m.cntsX + gx
}
