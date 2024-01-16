package aoigrid

import (
	"sync"

	"github.com/xhaoh94/gox/engine/types"
)

type (
	AOIGrid[T types.AOIKey] struct {
		GID     int
		Left    float32      //格子左边界坐标
		Right   float32      //格子右边界坐标
		Top     float32      //格子上边界坐标
		Bottom  float32      //格子下边界坐标
		ids     map[T]bool   //当前格子内的玩家或者物体成员ID
		uIDLock sync.RWMutex //playerIDs的保护map的锁
	}
)

// 初始化一个格子
func newGrid[T types.AOIKey](gID int, left, right, top, bottom float32) *AOIGrid[T] {
	return &AOIGrid[T]{
		GID:    gID,
		Left:   left,
		Right:  right,
		Top:    top,
		Bottom: bottom,
		ids:    make(map[T]bool),
	}
}

// 向当前格子中添加一个玩家
func (g *AOIGrid[T]) Add(id T) {
	g.uIDLock.Lock()
	defer g.uIDLock.Unlock()
	g.ids[id] = true
}

// 从格子中删除一个玩家
func (g *AOIGrid[T]) Remove(id T) {
	g.uIDLock.Lock()
	defer g.uIDLock.Unlock()
	delete(g.ids, id)
}

// 得到当前格子中所有的玩家
func (g *AOIGrid[T]) GetIDs() (ids []T) {
	g.uIDLock.RLock()
	defer g.uIDLock.RUnlock()
	for k := range g.ids {
		ids = append(ids, k)
	}
	return
}
