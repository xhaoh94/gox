package aoilink

import (
	"math"
	"sync"

	"github.com/xhaoh94/gox/engine/app"
	"github.com/xhaoh94/gox/engine/helper/mathhelper"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	AOILinkManager[T types.AOIKey] struct {
		Distance float32 //搜索范围距离
		lock     sync.RWMutex
		nodes    map[T]*AOINode[T]
		xLink    *AOILink[T]
		yLink    *AOILink[T]
	}
)

func (m *AOILinkManager[T]) Init() {
	m.nodes = make(map[T]*AOINode[T], 0)
	m.xLink = newAOILink[T](xLink)
	m.yLink = newAOILink[T](yLink)
}

func (m *AOILinkManager[T]) Enter(id T, x, y float32) {
	defer app.Recover()
	node := &AOINode[T]{
		id: id,
	}
	node.x, node.y = x, y
	m.xLink.Insert(node)
	m.yLink.Insert(node)
	m.lock.Lock()
	m.nodes[id] = node
	m.lock.Unlock()
}

func (m *AOILinkManager[T]) Leave(id T) {
	defer app.Recover()
	m.lock.RLock()
	node, ok := m.nodes[id]
	m.lock.RUnlock()
	if ok {
		m.xLink.Remove(node)
		m.yLink.Remove(node)
		m.lock.Lock()
		delete(m.nodes, id)
		m.lock.Unlock()
	}
}

func (m *AOILinkManager[T]) Update(id T, x, y float32) {
	defer app.Recover()
	m.lock.RLock()
	node, ok := m.nodes[id]
	m.lock.RUnlock()
	if ok {
		oldX := node.x
		oldY := node.y
		node.x, node.y = x, y
		if oldX != x {
			m.xLink.Update(node, oldX)
		}
		if oldY != y {
			m.yLink.Update(node, oldY)
		}
	}
}

func (m *AOILinkManager[T]) Find(id T) types.IAOIResult[T] {
	defer app.Recover()
	m.lock.RLock()
	node, ok := m.nodes[id]
	m.lock.RUnlock()
	result := newResult[T](id)
	if ok {
		for i := 0; i < 2; i++ {
			var xNode *AOINode[T]
			if i == 0 {
				xNode = node.xNext
			} else {
				xNode = node.xPrev
			}
			for xNode != nil {
				if math.Abs(float64(xNode.x-node.x)) > float64(m.Distance) {
					break
				}
				if mathhelper.Distance(node.x, node.y, xNode.x, xNode.y) <= m.Distance {
					result.push(xNode.id)
				}
				if i == 0 {
					xNode = xNode.xNext
				} else {
					xNode = xNode.xPrev
				}
			}

			var yNode *AOINode[T]
			if i == 0 {
				yNode = node.yNext
			} else {
				yNode = node.yPrev
			}
			for yNode != nil {
				if math.Abs(float64(yNode.y-node.y)) > float64(m.Distance) {
					break
				}
				if !result.Has(yNode.id) && mathhelper.Distance(node.x, node.y, yNode.x, yNode.y) <= m.Distance {
					result.push(yNode.id)
				}
				if i == 0 {
					yNode = yNode.yNext
				} else {
					yNode = yNode.yPrev
				}
			}
		}
	}
	result.push(id)
	return result
}
