package aoilink

import "github.com/xhaoh94/gox/engine/types"

const (
	xLink = 1
	yLink = 2
)

type (
	AOILink[T types.AOIKey] struct {
		linkType int
		count    int
		head     *AOINode[T]
		tail     *AOINode[T]
	}
)

func newAOILink[T types.AOIKey](linkType int) *AOILink[T] {
	return &AOILink[T]{linkType: linkType}
}
func (link *AOILink[T]) Insert(node *AOINode[T]) {
	insertPos := node.getValue(link.linkType)
	if link.head != nil {
		p := link.head
		for p != nil && p.getValue(link.linkType) < insertPos {
			p = p.GetNext(link.linkType)
		}

		if p == nil {
			tail := link.tail
			tail.setNext(link.linkType, node)
			node.setPrev(link.linkType, tail)
			link.tail = node
		} else {
			prev := p.getPrev(link.linkType)
			node.setNext(link.linkType, p)
			p.setPrev(link.linkType, node)
			node.setPrev(link.linkType, prev)
			if prev != nil {
				prev.setNext(link.linkType, node)
			} else {
				link.head = node
			}
		}
	} else {
		link.head = node
		link.tail = node
	}
	link.count++
}
func (link *AOILink[T]) Remove(node *AOINode[T]) {
	prev := node.getPrev(link.linkType)
	next := node.GetNext(link.linkType)
	if prev != nil {
		prev.setNext(link.linkType, next)
		node.setPrev(link.linkType, nil)
	} else {
		link.head = next
	}
	if next != nil {
		next.setPrev(link.linkType, prev)
		node.setNext(link.linkType, nil)
	} else {
		link.tail = prev
	}
	link.count--
}
func (link *AOILink[T]) Update(node *AOINode[T], oldPos float32) {
	pos := node.getValue(link.linkType)
	if pos > oldPos {
		next := node.GetNext(link.linkType)
		if next == nil || next.getValue(link.linkType) >= pos {
			return
		}
		prev := node.getPrev(link.linkType)
		if prev != nil {
			prev.setNext(link.linkType, next)
		} else {
			link.head = next
		}
		next.setPrev(link.linkType, prev)

		prev, next = next, next.GetNext(link.linkType)
		for next != nil && next.getValue(link.linkType) < pos {
			prev, next = next, next.GetNext(link.linkType)
		}
		prev.setNext(link.linkType, node)
		node.setPrev(link.linkType, prev)

		if next != nil {
			next.setPrev(link.linkType, node)
		} else {
			link.tail = node
		}
		node.setNext(link.linkType, next)

	} else {
		prev := node.getPrev(link.linkType)
		if prev == nil || prev.getValue(link.linkType) <= pos {
			return
		}

		next := node.GetNext(link.linkType)
		if next != nil {
			next.setPrev(link.linkType, prev)
		} else {
			link.tail = prev
		}
		prev.setNext(link.linkType, next)

		next, prev = prev, prev.getPrev(link.linkType)
		for prev != nil && prev.getValue(link.linkType) > pos {
			next, prev = prev, prev.getPrev(link.linkType)
		}
		next.setPrev(link.linkType, node)
		node.setNext(link.linkType, next)
		if prev != nil {
			prev.setNext(link.linkType, node)
		} else {
			link.head = node
		}
		node.setPrev(link.linkType, prev)
	}
}
