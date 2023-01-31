package aoigrid

import (
	"fmt"
	"math"
	"sync"

	"github.com/xhaoh94/gox/engine/aoi/aoibase"
)

type (
	AOIGridManager struct {
		Left       int //区域左边界坐标
		Right      int //区域右边界坐标
		Top        int //区域上边界坐标
		Bottom     int //区域下边界坐标
		GridWidth  int //格子的宽
		GridHeight int //格子的高
		Distance   int //搜索范围距离 ->当前周围 n格子范围

		cntsX      int              //x方向格子的数量
		cntsY      int              //y方向的格子数量
		grids      map[int]*AOIGrid //当前区域中都有哪些格子，key=格子ID， value=格子对象
		idToGridID map[string]int   //物体对应的格子
		lock       sync.RWMutex
	}
)

func (m *AOIGridManager) Init() {
	m.grids = make(map[int]*AOIGrid, 0)
	m.idToGridID = make(map[string]int, 0)
	m.cntsX = int(math.Ceil(float64(m.Right-m.Left) / float64(m.GridWidth)))
	m.cntsY = int(math.Ceil(float64(m.Bottom-m.Top) / float64(m.GridHeight)))
	//给AOI初始化区域中所有的格子
	for y := 0; y < m.cntsY; y++ {
		for x := 0; x < m.cntsX; x++ {
			//计算格子ID
			//格子编号：id = idy *nx + idx  (利用格子坐标得到格子编号)
			gid := y*m.cntsX + x
			//初始化一个格子放在AOI中的map里，key是当前格子的ID
			minx := m.Left + x*m.GridWidth
			maxx := m.Left + (x+1)*m.GridWidth
			if maxx > m.Right {
				maxx = m.Right
			}
			miny := m.Top + y*m.GridHeight
			maxy := m.Left + (y+1)*m.GridHeight
			if maxy > m.Bottom {
				maxy = m.Bottom
			}
			m.grids[gid] = newGrid(gid, minx, maxx, miny, maxy)
		}
	}
}
func (m *AOIGridManager) Enter(id string, x, y float32) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if gID, ok := m.idToGridID[id]; ok {
		m.grids[gID].Remove(id)
	}
	gID := m.getGridIDByPos(x, y)
	m.grids[gID].Add(id)
	m.idToGridID[id] = gID
}
func (m *AOIGridManager) Leave(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if gID, ok := m.idToGridID[id]; ok {
		m.grids[gID].Remove(id)
		delete(m.idToGridID, id)
	}
}
func (m *AOIGridManager) Update(id string, x, y float32) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if oldID, ok := m.idToGridID[id]; ok {
		newID := m.getGridIDByPos(x, y)
		if oldID != newID {
			m.grids[oldID].Remove(id)
			m.grids[newID].Add(id)
			m.idToGridID[id] = newID
		}
	}
}

func (m *AOIGridManager) Find(id string) aoibase.IAOIResult {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ids := newResult()
	if gID, ok := m.idToGridID[id]; ok {
		grids := m.getSurroundGridsByGid(gID)
		for i := range grids {
			v := grids[i].GetIDs()
			for j := range v {
				ids.push(v[j])
			}
		}
	}
	return ids
}

// 通过横纵坐标得到周边九宫格内的全部PlayerIDs
func (m *AOIGridManager) FindByPos(x, y float32) (ids []string) {
	//根据横纵坐标得到当前坐标属于哪个格子ID
	gID := m.getGridIDByPos(x, y)
	//根据格子ID得到周边九宫格的信息
	grids := m.getSurroundGridsByGid(gID)
	for i := range grids {
		v := grids[i]
		ids = append(ids, v.GetIDs()...)
	}
	return
}

// 根据格子的gID得到当前周边的九宫格信息
func (m *AOIGridManager) getSurroundGridsByGid(gID int) (grids []*AOIGrid) {
	//判断gID是否存在
	if _, ok := m.grids[gID]; !ok {
		return
	}
	idx := gID % m.cntsX
	idy := gID / m.cntsX
	for y := idy - m.Distance; y <= idy+m.Distance; y++ {
		if y < 0 || y >= m.cntsY {
			continue
		}
		for x := idx - m.Distance; x <= idx+m.Distance; x++ {
			if x < 0 || x >= m.cntsX {
				continue
			}
			id := gID + (x - idx) + ((y - idy) * m.cntsX)
			grids = append(grids, m.grids[id])
		}
	}
	return
}

// 通过横纵坐标获取对应的格子ID
func (m *AOIGridManager) getGridIDByPos(x, y float32) int {
	gx := int(x-float32(m.Left)) / m.GridWidth
	gy := int(y-float32(m.Top)) / m.GridHeight
	return gy*m.cntsX + gx
}

// 打印信息方法
func (m *AOIGridManager) String() string {
	s := fmt.Sprintf("AOIManagr:\nminX:%d, maxX:%d, cntsX:%d, minY:%d, maxY:%d, cntsY:%d\n Grids in AOI Manager:\n",
		m.Left, m.Right, m.cntsX, m.Top, m.Bottom, m.cntsY)
	for k := range m.grids {
		s += fmt.Sprintln(m.grids[k])
	}
	return s
}
