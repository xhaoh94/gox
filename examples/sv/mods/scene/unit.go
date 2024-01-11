package scene

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
)

type (
	Unit struct {
		location.Location
		Entity      *pb.Entity
		GateSession types.ISession
		Scene       *Scene
	}
)

func newUnit(entity *pb.Entity, gateSession types.ISession, scene *Scene) *Unit {
	unit := &Unit{Entity: entity, GateSession: gateSession, Scene: scene}
	gox.Location.Add(unit) //添加到Location
	return unit
}

func (unit *Unit) LocationID() uint32 {
	return unit.Entity.RoleId
}

func (unit *Unit) OnInit() {
	protoreg.AddLocation(unit, unit.LevaeMap)
	protoreg.AddLocation(unit, unit.Move)
}

func (unit *Unit) LevaeMap(ctx context.Context, session types.ISession, req *pb.C2S_LeaveMap) {
	protoreg.RemoveLocation(unit)
	logger.Debug().Msgf("玩家离开roleID:%d", unit.Entity.RoleId)
	unit.Scene.interiorLeaveVision(unit.Entity)
}

func (unit *Unit) Move(ctx context.Context, session types.ISession, req *pb.C2S_Move) {
	logger.Debug().Msgf("玩家移动roleID:%d", unit.Entity.RoleId)
	unit.Entity.Position = req.Points[len(req.Points)-1]
	unit.Scene.interiorMove(unit.Entity, req.Points)
}
