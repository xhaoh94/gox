package actor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/network/cmdtool"
	"github.com/xhaoh94/gox/engine/network/sv"
	"github.com/xhaoh94/gox/engine/rpc"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
)

var (
	actorPool *sync.Pool
)

func init() {
	actorPool = &sync.Pool{
		New: func() interface{} {
			return ActorConf{}
		},
	}
}
func New(engine types.IEngine) *ActorCrtl {
	return &ActorCrtl{
		engine:         engine,
		actorPrefix:    "location/actor",
		keyToActorConf: make(map[string]ActorConf),
	}
}

type (
	ActorCrtl struct {
		engine         types.IEngine
		actorPrefix    string
		es             *etcd.EtcdService
		keyLock        sync.RWMutex
		keyToActorConf map[string]ActorConf
	}
	ActorConf struct {
		ActorID   uint32
		ServiceID uint
	}
	Actor struct {
		fnLock  sync.RWMutex
		fnList  []interface{}
		cmdLock sync.RWMutex
		cmdList []uint32
	}
)

//AddActorFn 添加Actor回调
func (art *Actor) AddActorFn(fn interface{}) {
	defer art.fnLock.Unlock()
	art.fnLock.Lock()
	if art.fnList == nil {
		art.fnList = make([]interface{}, 0)
	}
	art.fnList = append(art.fnList, fn)
}

func (art *Actor) Destroy() {
	art.fnList = nil
	art.cmdList = nil
}

func (art *Actor) GetFnList() []interface{} {
	defer art.fnLock.RUnlock()
	art.fnLock.RLock()
	return art.fnList
}

func (art *Actor) GetCmdList() []uint32 {
	defer art.cmdLock.RUnlock()
	art.cmdLock.RLock()
	return art.cmdList
}
func (art *Actor) SetCmdList(cmd uint32) {
	defer art.cmdLock.Unlock()
	art.cmdLock.Lock()
	if art.cmdList == nil {
		art.cmdList = make([]uint32, 0)
	}
	art.cmdList = append(art.cmdList, cmd)
}

func (crtl *ActorCrtl) parseFn(aid uint32, fn interface{}) uint32 {
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
		break
	default:
		xlog.Error("Actor回调参数有误")
		break
	}
	var in reflect.Type
	switch tFun.NumIn() {
	case 1: //ctx
		break
	case 2: //ctx,req 或 ctx,session
		if out != nil {
			in = tFun.In(1)
		}
		break
	case 3: // ctx,session,req
		if tFun.NumOut() == 1 {
			xlog.Error("Actor回调参数有误")
			return 0
		}
		in = tFun.In(2)
		break
	default:
		xlog.Error("Actor回调参数有误")
		break
	}
	if out == nil && in == nil {
		xlog.Error("Actor回调参数有误")
		return 0
	}
	cmd := cmdtool.ToCmdByRtype(in, out, aid)
	if cmd == 0 {
		xlog.Error("Actor转换cmd错误")
		return cmd
	}
	if in != nil {
		crtl.engine.GetNetWork().RegisterRType(cmd, in)
	}
	crtl.engine.GetEvent().Bind(cmd, fn)
	return cmd
}

func (crtl *ActorCrtl) Add(actor types.IActor) {
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
	reg := &ActorConf{ActorID: aid, ServiceID: crtl.engine.ServiceID()}
	b, err := json.Marshal(reg)
	if err != nil {
		xlog.Error("Actor解析Json错误[%v]", err)
		return
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := crtl.parseFn(aid, fn); cmd != 0 {
			actor.SetCmdList(cmd)
		}
	}

	key := fmt.Sprintf(crtl.actorPrefix+"/%d", aid)
	crtl.es.Put(key, string(b))
}
func (crtl *ActorCrtl) Del(actor types.IActor) {
	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	cmdList := actor.GetCmdList()
	if cmdList != nil {
		for index := range cmdList {
			cmd := cmdList[index]
			crtl.engine.GetNetWork().UnRegisterRType(cmd)
			crtl.engine.GetEvent().UnBind(cmd)
		}
	}

	key := fmt.Sprintf(crtl.actorPrefix+"/%d", aid)
	if _, ok := crtl.keyToActorConf[key]; ok {
		crtl.es.Del(key)
	}
	actor.Destroy()
}
func (crtl *ActorCrtl) Get(actorID uint32) (ActorConf, bool) {
	defer crtl.keyLock.RUnlock()
	crtl.keyLock.RLock()
	key := fmt.Sprintf(crtl.actorPrefix+"/%d", actorID)
	actor, ok := crtl.keyToActorConf[key]
	if !ok {
		xlog.Error("找不到对应的Actor[%d]", actorID)
		return ActorConf{}, false
	}
	return actor, true
}
func (crtl *ActorCrtl) getSession(actorID uint32) *sv.Session {
	conf, ok := crtl.Get(actorID)
	if !ok {
		return nil
	}
	svConf := crtl.engine.GetNetWork().GetServiceCtrl().GetServiceConfByID(conf.ServiceID)
	if svConf == nil {
		xlog.Error("Actor没有找到服务 ServiceID:[%s]", conf.ServiceID)
		return nil
	}
	session := crtl.engine.GetNetWork().GetSessionByAddr(svConf.GetInteriorAddr())
	if session == nil {
		xlog.Error("Actor没有找到session[%d]", svConf.GetInteriorAddr())
		return nil
	}
	return session.(*sv.Session)
}

func (crtl *ActorCrtl) Send(actorID uint32, msg interface{}) bool {
	session := crtl.getSession(actorID)
	if session == nil {
		return false
	}
	cmd := cmdtool.ToCmd(msg, nil, actorID)
	return session.Send(cmd, msg)
}
func (crtl *ActorCrtl) Call(actorID uint32, msg interface{}, response interface{}) types.IDefaultRPC {
	session := crtl.getSession(actorID)
	if session == nil {
		dr := rpc.NewDefaultRpc(0, context.TODO(), response)
		defer dr.Run(false)
		return dr
	}
	return session.ActorCall(actorID, msg, response)
}
