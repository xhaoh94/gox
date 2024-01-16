package aoilink

import "github.com/xhaoh94/gox/engine/types"

type (
	AOINode[T types.AOIKey] struct {
		id           T
		x            float32
		y            float32
		xPrev, xNext *AOINode[T]
		yPrev, yNext *AOINode[T]
	}
)

func (node *AOINode[T]) getValue(linkType int) float32 {
	switch linkType {
	case xLink:
		return node.x
	case yLink:
		return node.y
	}
	return 0
}

func (node *AOINode[T]) getPrev(linkType int) *AOINode[T] {
	switch linkType {
	case xLink:
		return node.xPrev
	case yLink:
		return node.yPrev
	}
	return nil
}
func (node *AOINode[T]) setPrev(linkType int, aoi *AOINode[T]) {
	switch linkType {
	case xLink:
		node.xPrev = aoi
	case yLink:
		node.yPrev = aoi
	}
}
func (node *AOINode[T]) GetNext(linkType int) *AOINode[T] {
	switch linkType {
	case xLink:
		return node.xNext
	case yLink:
		return node.yNext
	}
	return nil
}
func (node *AOINode[T]) setNext(linkType int, aoi *AOINode[T]) {
	switch linkType {
	case xLink:
		node.xNext = aoi
	case yLink:
		node.yNext = aoi
	}
}
