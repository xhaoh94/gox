package aoilink

import (
	"math"
	"sync"

	"github.com/xhaoh94/gox/engine/aoi/aoibase"
	"github.com/xhaoh94/gox/util"
)

type (
	AOILinkManager struct {
		Distance float32 //搜索范围距离
		lock     sync.RWMutex
		nodes    map[string]*AOINode
		xLink    *AOILink
		yLink    *AOILink
	}
)

func (m *AOILinkManager) Init() {
	m.nodes = make(map[string]*AOINode, 0)
	m.xLink = newAOILink(xLink)
	m.yLink = newAOILink(yLink)
}

func (m *AOILinkManager) Enter(id string, x, y float32) {
	m.lock.Lock()
	defer m.lock.Unlock()
	node := &AOINode{
		id: id,
	}
	node.x, node.y = x, y
	m.xLink.Insert(node)
	m.yLink.Insert(node)
	m.nodes[id] = node
}

func (m *AOILinkManager) Leave(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if node, ok := m.nodes[id]; ok {
		m.xLink.Remove(node)
		m.yLink.Remove(node)
		delete(m.nodes, id)
	}
}

func (m *AOILinkManager) Update(id string, x, y float32) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if node, ok := m.nodes[id]; ok {
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

func (m *AOILinkManager) Find(id string) aoibase.IAOIResult {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result := newResult()
	if node, ok := m.nodes[id]; ok {
		for i := 0; i < 2; i++ {
			var xNode *AOINode
			if i == 0 {
				xNode = node.xNext
			} else {
				xNode = node.xPrev
			}
			for xNode != nil {
				if math.Abs(float64(xNode.x-node.x)) > float64(m.Distance) {
					break
				}
				if util.Distance(node.x, node.y, xNode.x, xNode.y) <= m.Distance {
					result.push(xNode.id)
				}
				if i == 0 {
					xNode = xNode.xNext
				} else {
					xNode = xNode.xPrev
				}
			}

			var yNode *AOINode
			if i == 0 {
				yNode = node.yNext
			} else {
				yNode = node.yPrev
			}
			for yNode != nil {
				if math.Abs(float64(yNode.y-node.y)) > float64(m.Distance) {
					break
				}
				if !result.get(yNode.id) && util.Distance(node.x, node.y, yNode.x, yNode.y) <= m.Distance {
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
