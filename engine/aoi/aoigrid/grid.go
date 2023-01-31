package aoigrid

import (
	"fmt"
	"sync"
)

type (
	AOIGrid struct {
		GID     int
		Left    int             //格子左边界坐标
		Right   int             //格子右边界坐标
		Top     int             //格子上边界坐标
		Bottom  int             //格子下边界坐标
		ids     map[string]bool //当前格子内的玩家或者物体成员ID
		uIDLock sync.RWMutex    //playerIDs的保护map的锁
	}
)

//初始化一个格子
func newGrid(gID, left, right, top, bottom int) *AOIGrid {
	return &AOIGrid{
		GID:    gID,
		Left:   left,
		Right:  right,
		Top:    top,
		Bottom: bottom,
		ids:    make(map[string]bool),
	}
}

//向当前格子中添加一个玩家
func (g *AOIGrid) Add(id string) {
	g.uIDLock.Lock()
	defer g.uIDLock.Unlock()
	g.ids[id] = true
}

//从格子中删除一个玩家
func (g *AOIGrid) Remove(id string) {
	g.uIDLock.Lock()
	defer g.uIDLock.Unlock()
	delete(g.ids, id)
}

//得到当前格子中所有的玩家
func (g *AOIGrid) GetIDs() (ids []string) {
	g.uIDLock.RLock()
	defer g.uIDLock.RUnlock()
	for k := range g.ids {
		ids = append(ids, k)
	}
	return
}

//打印信息方法
func (g *AOIGrid) String() string {
	return fmt.Sprintf("Grid id: %d, minX:%d, maxX:%d, minY:%d, maxY:%d, playerIDs:%v",
		g.GID, g.Left, g.Right, g.Top, g.Bottom, g.ids)
}
