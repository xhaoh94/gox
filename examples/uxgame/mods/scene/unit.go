package scene

import (
	"context"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/common/vec"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/mgrs/timemgr"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/examples/pb"
)

type (
	Unit struct {
		location.Location
		Entity      *pb.Entity
		GateSession uint32
		Scene       *Scene

		AOIResult types.IAOIResult[uint32]

		isMove    bool
		moveIndex int
		points    []*pb.Vector3
	}
)

func newUnit(entity *pb.Entity, gateSession types.ISession) *Unit {
	unit := &Unit{
		Entity:      entity,
		GateSession: gateSession.ID(),
	}
	gox.Location.Register(unit) //添加到Location
	timemgr.Add(unit.update, true)
	return unit
}

func (unit *Unit) LocationID() uint32 {
	return unit.Entity.RoleId
}

func (unit *Unit) Position() vec.Vector3 {
	return vec.CreateVector3(unit.Entity.Position.X, unit.Entity.Position.Y, unit.Entity.Position.Z)
}
func (unit *Unit) SetPosition(pos vec.Vector3) {
	unit.Entity.Position.X = pos.X
	unit.Entity.Position.Y = pos.Y
	unit.Entity.Position.Z = pos.Z
}

// 转AOI坐标，由于AOI格子模式由左上角，需要将对应的坐标转换到AOI里面
func (unit *Unit) AOIPosition() vec.Vector2 {
	//
	x := unit.Entity.Position.X + float32(unit.Scene.Width)/2
	y := unit.Entity.Position.Z + float32(unit.Scene.Height)/2
	return vec.CreateVector2(x, y)
}

func (unit *Unit) update() {
	unit.move()
}
func (unit *Unit) move() {
	if !unit.isMove {
		return
	}
	cnt := len(unit.points)
	if cnt == 0 {
		return
	}
	if unit.moveIndex < cnt {
		point := unit.points[unit.moveIndex]
		target := vec.CreateVector3(point.X, point.Y, point.Z)
		nowPos := unit.Position()
		dir := target.Sub(nowPos)
		unit.SetPosition(nowPos.Add(dir.Normalize().MulNumber(timemgr.DeltaTime * 5)))
		unit.freshAOI()
		if dir.SqrMagnitude() <= 0.1 {
			unit.moveIndex++
		}
	} else {
		unit.isMove = false
		clear(unit.points)
		unit.moveIndex = 0
	}
}
func (unit *Unit) freshAOI() {
	unit.Scene.aoiMgr.Update(unit.Entity.RoleId, unit.Position().X, unit.Position().Z)
	oldAOIResult := unit.AOIResult
	unit.AOIResult = unit.Scene.aoiMgr.Find(unit.Entity.RoleId)
	//获取AOI的补集、差集、交集
	Complement, Minus, Intersect := unit.AOIResult.Compare(oldAOIResult)
	//通知补集里的对象，玩家进入视野，并且玩家正在行走
	unit.Scene.interiorUnitIntoView(Complement, unit)
	unit.Scene.interiorUnitMove(Complement, unit, unit.points, unit.moveIndex)
	//通知差集里的对象，玩家离开视野
	unit.Scene.interiorUnitOutofView(Minus, unit)
	//通知交集的对象，玩家更新位置
	unit.Scene.interiorUnitPosition(Intersect, unit)
}

func (unit *Unit) OnInit() {
	protoreg.AddLocation(unit, unit.LevaeScene)
	protoreg.AddLocation(unit, unit.Move)
}

func (unit *Unit) EnterScene(scene *Scene) {
	unit.Scene = scene
	unit.Scene.aoiMgr.Enter(unit.Entity.RoleId, unit.Entity.Position.X, unit.Entity.Position.Z)
	unit.AOIResult = unit.Scene.aoiMgr.Find(unit.Entity.RoleId)
}

func (unit *Unit) LevaeScene(ctx context.Context, session types.ISession, req *pb.C2S_LeaveScene) {
	gox.Location.UnRegister(unit)
	logger.Debug().Msgf("玩家离开roleID:%d", unit.Entity.RoleId)
	unit.AOIResult = unit.Scene.aoiMgr.Find(unit.Entity.RoleId)
	unit.Scene.aoiMgr.Leave(unit.Entity.RoleId)
	unit.Scene.OnLeaveScene(unit)
}

func (unit *Unit) Move(ctx context.Context, session types.ISession, req *pb.C2S_Move) {
	unit.points = req.Points
	unit.isMove = true
	unit.moveIndex = 0
	oldAOIResult := unit.AOIResult
	unit.AOIResult = unit.Scene.aoiMgr.Find(unit.Entity.RoleId)
	//获取AOI的补集、差集、交集
	Complement, Minus, Intersect := unit.AOIResult.Compare(oldAOIResult)
	//通知补集里的对象，玩家进入视野
	unit.Scene.interiorUnitIntoView(Complement, unit)
	//通知差集里的对象，玩家离开视野
	unit.Scene.interiorUnitOutofView(Minus, unit)
	//通知补集和交集的对象，玩家开始行走
	ids := append(Complement, Intersect...)
	unit.Scene.interiorUnitMove(ids, unit, unit.points, unit.moveIndex)
}
