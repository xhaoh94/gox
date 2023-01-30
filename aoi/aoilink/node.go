package aoilink

type (
	AOINode struct {
		id           string
		x            float32
		y            float32
		xPrev, xNext *AOINode
		yPrev, yNext *AOINode
	}
)

func (node *AOINode) getValue(linkType int) float32 {
	switch linkType {
	case xLink:
		return node.x
	case yLink:
		return node.y
	}
	return 0
}

func (node *AOINode) getPrev(linkType int) *AOINode {
	switch linkType {
	case xLink:
		return node.xPrev
	case yLink:
		return node.yPrev
	}
	return nil
}
func (node *AOINode) setPrev(linkType int, aoi *AOINode) {
	switch linkType {
	case xLink:
		node.xPrev = aoi
	case yLink:
		node.yPrev = aoi
	}
}
func (node *AOINode) GetNext(linkType int) *AOINode {
	switch linkType {
	case xLink:
		return node.xNext
	case yLink:
		return node.yNext
	}
	return nil
}
func (node *AOINode) setNext(linkType int, aoi *AOINode) {
	switch linkType {
	case xLink:
		node.xNext = aoi
	case yLink:
		node.yNext = aoi
	}
}
