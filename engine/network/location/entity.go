package location

import (
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	Entity struct {
	}
)

func (*Entity) Init(entity types.ILocationEntity) {
	entity.OnInit()
}
func (*Entity) Destroy(entity types.ILocationEntity) {
	protoreg.RemoveLocation(entity)
}
