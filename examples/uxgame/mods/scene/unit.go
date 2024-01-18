package scene

import (
	"context"
	"sync"
	"time"

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
		UnitID   uint32
		UnitMask int32
		Position vec.Vector3

		GateSession uint32
		Scene       *Scene

		aoiResult types.IAOIResult[uint32]

		updatePositionTime time.Time
		moveMux            sync.RWMutex
		isMove             bool
		moveIndex          int
		points             []*pb.Vector3
	}
)

func newUnit(entity *pb.Entity, gateSession types.ISession) *Unit {
	unit := &Unit{
		UnitID:      entity.RoleId,
		UnitMask:    entity.RoleMask,
		Position:    vec.CreateVector3(entity.Position.X, entity.Position.Y, entity.Position.Z),
		GateSession: gateSession.ID(),
	}
	gox.Location.Register(unit) //添加到Location
	timemgr.Add(unit.update)
	return unit
}

func (unit *Unit) LocationID() uint32 {
	return unit.UnitID
}

func (unit *Unit) SetAOIResult(result types.IAOIResult[uint32]) {
	if unit.aoiResult != nil {
		unit.aoiResult.Reset()
	}
	unit.aoiResult = result
}
func (unit *Unit) GetAOIResult() types.IAOIResult[uint32] {
	return unit.aoiResult
}

func (unit *Unit) update() {
	unit.move()
}
func (unit *Unit) move() {
	unit.moveMux.Lock()
	defer unit.moveMux.Unlock()
	if !unit.isMove {
		return
	}
	if unit.points == nil {
		return
	}
	if unit.moveIndex < len(unit.points) {
		point := unit.points[unit.moveIndex]
		target := vec.CreateVector3(point.X, point.Y, point.Z)
		dir := target.Sub(unit.Position)
		unit.Position = unit.Position.Add(dir.Normalize().MulNumber(timemgr.DeltaTime * 5))
		unit.freshAOI()
		if dir.SqrMagnitude() <= 0.1 {
			unit.moveIndex++
		}
	} else {
		unit.isMove = false
		unit.points = nil
		unit.moveIndex = 0
	}
}
func (unit *Unit) freshAOI() {
	unit.Scene.aoiMgr.Update(unit.UnitID, unit.Position.X, unit.Position.Z)
	oldAOIResult := unit.GetAOIResult()
	newAOIResult := unit.Scene.aoiMgr.Find(unit.UnitID)
	//获取AOI的补集、差集、交集
	Complement, Minus, Intersect := newAOIResult.Compare(oldAOIResult)
	unit.SetAOIResult(newAOIResult)
	//通知补集里的对象，玩家进入视野，并且玩家正在行走
	unit.Scene.interiorUnitIntoView(Complement, unit)
	unit.Scene.interiorUnitMove(Complement, unit, unit.points, unit.moveIndex)
	//通知差集里的对象，玩家离开视野
	unit.Scene.interiorUnitOutofView(Minus, unit)

	//500毫秒广播更新一下位置
	if time.Since(unit.updatePositionTime) >= time.Millisecond*500 {
		//通知交集的对象，玩家更新位置
		unit.Scene.interiorUnitPosition(Intersect, unit)
		unit.updatePositionTime = time.Now()
	}
}

func (unit *Unit) OnInit() {
	protoreg.AddLocation(unit, unit.LevaeScene)
	protoreg.AddLocation(unit, unit.Move)
}

func (unit *Unit) EnterScene(scene *Scene) {
	unit.Scene = scene
	unit.Scene.aoiMgr.Enter(unit.UnitID, unit.Position.X, unit.Position.Z)
	unit.SetAOIResult(unit.Scene.aoiMgr.Find(unit.UnitID))
}

func (unit *Unit) LevaeScene(ctx context.Context, session types.ISession, req *pb.C2S_LeaveScene) {
	gox.Location.UnRegister(unit)
	logger.Debug().Msgf("玩家离开roleID:%d", unit.UnitID)
	unit.SetAOIResult(unit.Scene.aoiMgr.Find(unit.UnitID))
	unit.Scene.aoiMgr.Leave(unit.UnitID)
	unit.Scene.OnLeaveScene(unit)
}

func (unit *Unit) Move(ctx context.Context, session types.ISession, req *pb.C2S_Move) {
	unit.moveMux.Lock()
	defer unit.moveMux.Unlock()

	unit.updatePositionTime = time.Now()
	unit.points = req.Points
	unit.isMove = true
	unit.moveIndex = 0
	oldAOIResult := unit.GetAOIResult()
	newAOIResult := unit.Scene.aoiMgr.Find(unit.UnitID)
	//获取AOI的补集、差集、交集
	Complement, Minus, Intersect := newAOIResult.Compare(oldAOIResult)
	unit.SetAOIResult(newAOIResult)
	//通知补集里的对象，玩家进入视野
	unit.Scene.interiorUnitIntoView(Complement, unit)
	//通知差集里的对象，玩家离开视野
	unit.Scene.interiorUnitOutofView(Minus, unit)
	//通知补集和交集的对象，玩家开始行走
	ids := append(Complement, Intersect...)
	ids = append(ids, unit.UnitID)
	unit.Scene.interiorUnitMove(ids, unit, unit.points, unit.moveIndex)
}
