package location

import (
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	Location struct {
	}
)

func (*Location) Init(entity types.ILocation) {
	entity.OnInit()
}
func (*Location) Destroy(entity types.ILocation) {
	protoreg.RemoveLocation(entity)
}
