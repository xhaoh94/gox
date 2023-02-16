package location

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	LocationSystem struct {
		gox.Module
		SyncLocation
		context   context.Context
		wg        sync.WaitGroup
		lock      sync.RWMutex
		entitys   map[uint32]uint
		cacellock sync.RWMutex
		cacel     map[uint32]uint
	}
	LocationInit struct {
		Entitys []LocationEntity
	}
	LocationLock struct {
		Lock bool
	}
	LocationEntity struct {
		ActorID uint32
		AppID   uint
	}
	LocationReslut struct {
	}
)

func New(ctx context.Context) *LocationSystem {
	return &LocationSystem{
		context: ctx,
	}
}
func (m *LocationSystem) Init() {
	m.entitys = make(map[uint32]uint)
	m.cacel = make(map[uint32]uint)
	m.RegisterRpc(consts.LocationLock, m.LockHandler)
	m.RegisterRpc(consts.LocationRefresh, m.RefreshHandler)
	m.RegisterRpc(consts.LocationInit, m.InitHandler)
}
func (m *LocationSystem) Start() {
	m.SyncLocation.Lock()
	if entitys := m.SyncLocation.GetEntitys(); entitys != nil && len(entitys.Entitys) > 0 {
		m.lock.Lock()
		for _, v := range entitys.Entitys {
			m.entitys[v.ActorID] = v.AppID
		}
		defer m.lock.Unlock()
	}
	m.SyncLocation.UnLock()
}
func (m *LocationSystem) Stop() {

}

func (m *LocationSystem) InitHandler(ctx context.Context) *LocationInit {
	defer m.lock.RUnlock()
	m.lock.RLock()
	entitys := make([]LocationEntity, len(m.entitys))
	for aid, appId := range m.entitys {
		entitys = append(entitys, LocationEntity{ActorID: aid, AppID: appId})
	}
	return &LocationInit{Entitys: entitys}
}
func (m *LocationSystem) LockHandler(ctx context.Context, req *LocationLock) *LocationReslut {
	if req.Lock {
		m.wg.Add(1)
	} else {
		m.wg.Done()
	}
	return &LocationReslut{}
}
func (m *LocationSystem) RefreshHandler(ctx context.Context, req *LocationEntity) *LocationReslut {
	m.addOrdel(req.ActorID, req.AppID)
	return &LocationReslut{}
}
func (m *LocationSystem) addOrdel(actorID uint32, appID uint) {
	defer m.lock.Unlock()
	m.lock.Lock()
	if appID == 0 {
		delete(m.entitys, actorID)
	} else {
		m.entitys[actorID] = appID
	}
}
func (as *LocationSystem) getAppId(actorID uint32) uint {
	if cacelId, ok := as.cacel[actorID]; ok {
		return cacelId
	} else {
		as.wg.Wait()
		defer as.lock.RUnlock()
		as.lock.RLock()
		if id, ok := as.entitys[actorID]; ok && id != cacelId {
			as.cacellock.Lock()
			as.cacel[actorID] = id
			as.cacellock.Unlock()
			return id
		}
		return 0
	}

}
func (as *LocationSystem) getSession(appID uint) types.ISession {
	serviceEntity := gox.NetWork.GetServiceEntityByID(appID)
	if serviceEntity == nil {
		xlog.Error("Actor没有找到服务 ServiceID:[%s]", appID)
		return nil
	}
	session := gox.NetWork.GetSessionByAddr(serviceEntity.GetInteriorAddr())
	if session == nil {
		xlog.Error("Actor没有找到session[%d]", serviceEntity.GetInteriorAddr())
		return nil
	}
	return session
}
func (as *LocationSystem) Send(actorID uint32, msg interface{}) bool {
	if actorID == 0 {
		xlog.Error("ActorCall传入ActorID不能为空")
		return false
	}
	defer as.cacellock.RUnlock()
	as.cacellock.RLock()
	loopCnt := 0
	cmd := cmdhelper.ToCmd(msg, nil, actorID)
	for {
		loopCnt++
		if loopCnt > 5 {
			return false
		}
		if id := as.getAppId(actorID); id > 0 {
			if session := as.getSession(id); session != nil {
				if id == gox.AppConf.Eid {
					if _, err := cmdhelper.CallEvt(cmd, as.context, session, msg); err == nil {
						return true
					} else {
						xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
					}
				} else {
					return session.Send(cmd, msg)
				}
			}
		}
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
}
func (as *LocationSystem) Call(actorID uint32, msg interface{}, response interface{}) types.IRpcx {
	if actorID == 0 {
		xlog.Error("ActorCall传入ActorID不能为空")
		return rpc.NewEmptyRpcx()
	}

	defer as.cacellock.RUnlock()
	as.cacellock.RLock()
	loopCnt := 0
	var cacelId uint = 0
	cmd := cmdhelper.ToCmd(msg, response, actorID)
	for {
		loopCnt++
		if loopCnt > 5 {
			return rpc.NewEmptyRpcx()
		}
		if id, ok := as.cacel[actorID]; ok {
			if id == gox.AppConf.Eid {
				if response, err := cmdhelper.CallEvt(cmd, as.context, msg); err == nil {
					rpcx := rpc.NewRpcx(as.context, response)
					defer rpcx.Run(true)
					return rpcx
				} else {
					xlog.Warn("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
					cacelId = id
				}
			} else {
				if session := as.getSession(id); session != nil {
					return session.CallByCmd(cmd, msg, response)
				}
			}
		}

		as.lock.RLock()
		if id, ok := as.entitys[actorID]; ok && id != cacelId {
			as.cacellock.Lock()
			as.cacel[actorID] = id
			as.cacellock.Unlock()
			as.lock.Unlock()
			continue
		}
		as.lock.Unlock()
		time.Sleep(time.Millisecond * 500) //等待0.5秒
	}
}

func (as *LocationSystem) Add(actor types.ILocationEntity) {
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	fnList := actor.GetFnList()
	if fnList == nil {
		xlog.Error("Actor没有注册回调函数")
		return
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := as.parseFn(aid, fn); cmd != 0 {
			actor.SetCmdList(cmd)
		}
	}

	as.SyncLocation.Lock()
	as.SyncLocation.Add(aid, gox.AppConf.Eid)
	as.addOrdel(aid, gox.AppConf.Eid)
	as.SyncLocation.UnLock()
}
func (as *LocationSystem) Del(actor types.ILocationEntity) {
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	cmdList := actor.GetCmdList()
	for index := range cmdList {
		cmd := cmdList[index]
		protoreg.UnRegisterRType(cmd)
		gox.Event.UnBind(cmd)
	}
	actor.Destroy()

	as.SyncLocation.Lock()
	as.SyncLocation.Remove(aid)
	as.addOrdel(aid, 0)
	as.SyncLocation.UnLock()
}
func (as *LocationSystem) parseFn(aid uint32, fn interface{}) uint32 {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	if tFun.Kind() != reflect.Func {
		xlog.Error("Actor回调不是方法")
		return 0
	}
	var out reflect.Type
	switch tFun.NumOut() {
	case 0: //存在没有返回参数的情况
		break
	case 1:
		out = tFun.Out(0)
	default:
		xlog.Error("Actor回调参数有误")
	}
	var in reflect.Type
	switch tFun.NumIn() {
	case 1: //ctx
		break
	case 2: //ctx,req 或 ctx,session
		if out != nil {
			in = tFun.In(1)
		}
	case 3: // ctx,session,req
		if tFun.NumOut() == 1 {
			xlog.Error("Actor回调参数有误")
			return 0
		}
		in = tFun.In(2)
	default:
		xlog.Error("Actor回调参数有误")
	}
	if out == nil && in == nil {
		xlog.Error("Actor回调参数有误")
		return 0
	}
	cmd := cmdhelper.ToCmdByRtype(in, out, aid)
	if cmd == 0 {
		xlog.Error("Actor转换cmd错误")
		return cmd
	}
	if in != nil {
		protoreg.RegisterRType(cmd, in)
	}
	gox.Event.Bind(cmd, fn)
	return cmd
}
