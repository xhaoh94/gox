package actor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"
)

var (
	actorPool *sync.Pool
)

func init() {
	actorPool = &sync.Pool{
		New: func() interface{} {
			return &ActorConf{}
		},
	}
}
func New(engine types.IEngine) *ActorCrtl {
	return &ActorCrtl{
		engine:         engine,
		actorPrefix:    "location/actor",
		keyToActorConf: make(map[string]*ActorConf),
		// aidToActor:     make(map[uint32]*Actor),
	}
}

type (
	ActorCrtl struct {
		engine         types.IEngine
		actorPrefix    string
		actorEs        *etcd.EtcdService
		keyLock        sync.RWMutex
		keyToActorConf map[string]*ActorConf
	}
	ActorConf struct {
		ActorID   uint32
		ServiceID uint
	}
	Actor struct {
		ActorID uint32

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
	art.ActorID = 0
	art.fnList = nil
	art.cmdList = nil
}

func (art *Actor) getFnList() []interface{} {
	defer art.fnLock.RUnlock()
	art.fnLock.RLock()
	return art.fnList
}

func (art *Actor) getCmdList() []uint32 {
	defer art.cmdLock.RUnlock()
	art.cmdLock.RLock()
	return art.cmdList
}
func (art *Actor) setCmdList(cmd uint32) {
	defer art.cmdLock.Unlock()
	art.cmdLock.Lock()
	if art.cmdList == nil {
		art.cmdList = make([]uint32, 0)
	}
	art.cmdList = append(art.cmdList, cmd)
}

func (actorCrtl *ActorCrtl) parseFn(fn interface{}) uint32 {
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
	var key string
	if out != nil {
		if out.Kind() != reflect.Ptr {
			xlog.Error("Actor输出回调参数需要是指针类型")
			return 0
		}
		key = out.Elem().Name()
	}
	var cmd uint32
	if in != nil {
		if in.Kind() != reflect.Ptr {
			xlog.Error("Actor输入回调参数需要是指针类型")
			return 0
		}
		key = in.Elem().Name() + key
		cmd = util.StringToHash(key)
		actorCrtl.engine.GetNetWork().RegisterRType(cmd, in)
	}
	if cmd == 0 {
		cmd = util.StringToHash(key)
	}
	actorCrtl.engine.GetEvent().Bind(cmd, fn)
	return cmd
}

func (actorCrtl *ActorCrtl) Add(atr types.IActor) {
	actor := atr.(*Actor)
	aid := actor.ActorID
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	fnList := actor.getFnList()
	if fnList == nil {
		xlog.Error("Actor没有注册回调函数")
		return
	}
	reg := &ActorConf{ActorID: aid, ServiceID: actorCrtl.engine.ServiceID()}
	b, err := json.Marshal(reg)
	if err != nil {
		xlog.Error("Actor解析Json错误[%v]", err)
		return
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := actorCrtl.parseFn(fn); cmd != 0 {
			actor.setCmdList(cmd)
		}
	}
	key := fmt.Sprintf(actorCrtl.actorPrefix+"/%d", aid)
	actorCrtl.actorEs.Put(key, string(b))
}
func (actorCrtl *ActorCrtl) Del(atr types.IActor) {
	defer actorCrtl.keyLock.Unlock()
	actorCrtl.keyLock.Lock()
	actor := atr.(*Actor)
	aid := actor.ActorID
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	cmdList := actor.getCmdList()
	if cmdList != nil {
		for index := range cmdList {
			cmd := cmdList[index]
			actorCrtl.engine.GetNetWork().UnRegisterRType(cmd)
			actorCrtl.engine.GetEvent().UnBind(cmd)
		}
	}

	key := fmt.Sprintf(actorCrtl.actorPrefix+"/%d", aid)
	if _, ok := actorCrtl.keyToActorConf[key]; ok {
		actorCrtl.actorEs.Del(key)
	}
	actor.Destroy()
}
func (actorCrtl *ActorCrtl) Get(actorID uint32) *ActorConf {
	defer actorCrtl.keyLock.RUnlock()
	actorCrtl.keyLock.RLock()
	key := fmt.Sprintf(actorCrtl.actorPrefix+"/%d", actorID)
	actor, ok := actorCrtl.keyToActorConf[key]
	if !ok {
		xlog.Error("找不到对应的actor。id[%d]", actorID)
		return nil
	}
	return actor
}
func (actorCrtl *ActorCrtl) getSession(actorID uint32) types.ISession {
	ar := actorCrtl.Get(actorID)
	if ar == nil {
		return nil
	}
	svConf := actorCrtl.engine.GetNetWork().GetServiceCtrl().GetServiceConfByID(ar.ServiceID)
	if svConf == nil {
		xlog.Error("actor找不到服务 ServiceID:[%s]", ar.ServiceID)
		return nil
	}
	session := actorCrtl.engine.GetNetWork().GetSessionByAddr(svConf.GetInteriorAddr())
	if session == nil {
		xlog.Error("actor没有找到session。id[%d]", svConf.GetInteriorAddr())
		return nil
	}
	return session
}

func (actorCrtl *ActorCrtl) Send(actorID uint32, msg interface{}) bool {
	session := actorCrtl.getSession(actorID)
	if session == nil {
		return false
	}
	cmd := util.ToCmd(msg, nil)
	return session.Send(cmd, msg)
}
func (actorCrtl *ActorCrtl) Call(actorID uint32, msg interface{}, response interface{}) types.IDefaultRPC {
	session := actorCrtl.getSession(actorID)
	if session == nil {
		dr := rpc.NewDefaultRpc(0, context.TODO(), response)
		defer dr.Run(false)
		return dr
	}
	return session.Call(msg, response)
}
