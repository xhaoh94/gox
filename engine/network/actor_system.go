package network

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// 	"time"

// 	"github.com/xhaoh94/gox"
// 	"github.com/xhaoh94/gox/engine/etcd"
// 	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
// 	"github.com/xhaoh94/gox/engine/network/protoreg"
// 	"github.com/xhaoh94/gox/engine/network/rpc"
// 	"github.com/xhaoh94/gox/engine/types"
// 	"github.com/xhaoh94/gox/engine/xlog"

// 	"github.com/coreos/etcd/mvcc/mvccpb"
// )

// const ()

// type (
// 	ActorSystem struct {
// 		etcd.EtcdComponent
// 		context        context.Context
// 		actorPrefix    string
// 		es             *etcd.EtcdConf
// 		keyToActorConf map[string]ActorEntity
// 	}

// 	ActorEntity struct {
// 		ActorID uint32
// 		AppID   uint
// 	}
// )

// func newActorSystem(ctx context.Context) *ActorSystem {
// 	return &ActorSystem{
// 		context:        ctx,
// 		actorPrefix:    "location/actor",
// 		keyToActorConf: make(map[string]ActorEntity),
// 	}
// }

// func (as *ActorSystem) parseFn(aid uint32, fn interface{}) uint32 {
// 	tVlaue := reflect.ValueOf(fn)
// 	tFun := tVlaue.Type()
// 	if tFun.Kind() != reflect.Func {
// 		xlog.Error("Actor回调不是方法")
// 		return 0
// 	}
// 	var out reflect.Type
// 	switch tFun.NumOut() {
// 	case 0: //存在没有返回参数的情况
// 		break
// 	case 1:
// 		out = tFun.Out(0)
// 	default:
// 		xlog.Error("Actor回调参数有误")
// 	}
// 	var in reflect.Type
// 	switch tFun.NumIn() {
// 	case 1: //ctx
// 		break
// 	case 2: //ctx,req 或 ctx,session
// 		if out != nil {
// 			in = tFun.In(1)
// 		}
// 	case 3: // ctx,session,req
// 		if tFun.NumOut() == 1 {
// 			xlog.Error("Actor回调参数有误")
// 			return 0
// 		}
// 		in = tFun.In(2)
// 	default:
// 		xlog.Error("Actor回调参数有误")
// 	}
// 	if out == nil && in == nil {
// 		xlog.Error("Actor回调参数有误")
// 		return 0
// 	}
// 	cmd := cmdhelper.ToCmdByRtype(in, out, aid)
// 	if cmd == 0 {
// 		xlog.Error("Actor转换cmd错误")
// 		return cmd
// 	}
// 	if in != nil {
// 		protoreg.RegisterRType(cmd, in)
// 	}
// 	gox.Event.Bind(cmd, fn)
// 	return cmd
// }

// func (as *ActorSystem) Add(actor types.IActorEntity) {
// 	aid := actor.ActorID()
// 	if aid == 0 {
// 		xlog.Error("Actor没有初始化ID")
// 		return
// 	}
// 	fnList := actor.GetFnList()
// 	if fnList == nil {
// 		xlog.Error("Actor没有注册回调函数")
// 		return
// 	}
// 	reg := &ActorEntity{ActorID: aid, AppID: gox.AppConf.Eid}
// 	b, err := json.Marshal(reg)
// 	if err != nil {
// 		xlog.Error("Actor解析Json错误[%v]", err)
// 		return
// 	}
// 	for index := range fnList {
// 		fn := fnList[index]
// 		if cmd := as.parseFn(aid, fn); cmd != 0 {
// 			actor.SetCmdList(cmd)
// 		}
// 	}

// 	key := fmt.Sprintf(as.actorPrefix+"/%d", aid)
// 	as.es.Put(key, string(b))
// }
// func (as *ActorSystem) Del(actor types.IActorEntity) {
// 	defer as.RUnlock()
// 	as.RLock()
// 	aid := actor.ActorID()
// 	if aid == 0 {
// 		xlog.Error("Actor没有初始化ID")
// 		return
// 	}
// 	cmdList := actor.GetCmdList()
// 	for index := range cmdList {
// 		cmd := cmdList[index]
// 		protoreg.UnRegisterRType(cmd)
// 		gox.Event.UnBind(cmd)
// 	}

// 	key := fmt.Sprintf(as.actorPrefix+"/%d", aid)
// 	if _, ok := as.keyToActorConf[key]; ok {
// 		as.es.Del(key)
// 	}
// 	actor.Destroy()
// }
// func (as *ActorSystem) Get(actorID uint32) (ActorEntity, bool) {
// 	defer as.RUnlock()
// 	as.RLock()
// 	key := fmt.Sprintf(as.actorPrefix+"/%d", actorID)
// 	actor, ok := as.keyToActorConf[key]
// 	if !ok {
// 		xlog.Error("找不到对应的Actor[%d]", actorID)
// 		return ActorEntity{}, false
// 	}
// 	return actor, true
// }
// func (as *ActorSystem) getSession(entityConf ActorEntity) types.ISession {
// 	svConf := gox.ServiceSystem.GetServiceEntityByID(entityConf.AppID)
// 	if svConf == nil {
// 		xlog.Error("Actor没有找到服务 ServiceID:[%s]", entityConf.AppID)
// 		return nil
// 	}
// 	session := gox.NetWork.GetSessionByAddr(svConf.GetInteriorAddr())
// 	if session == nil {
// 		xlog.Error("Actor没有找到session[%d]", svConf.GetInteriorAddr())
// 		return nil
// 	}
// 	return session
// }
// func (as *ActorSystem) Send(actorID uint32, msg interface{}) bool {
// 	entityConf, ok := as.Get(actorID)
// 	if !ok {
// 		return false
// 	}
// 	session := as.getSession(entityConf)
// 	if session == nil {
// 		return false
// 	}
// 	cmd := cmdhelper.ToCmd(msg, nil, actorID)
// 	if entityConf.AppID == gox.AppConf.Eid {
// 		if _, err := cmdhelper.CallEvt(cmd, as.context, session, msg); err != nil {
// 			xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
// 		}
// 		return true
// 	}
// 	return session.Send(cmd, msg)
// }
// func (as *ActorSystem) Call(actorID uint32, msg interface{}, response interface{}) types.IRpcx {
// 	entityConf, ok := as.Get(actorID)
// 	if !ok {
// 		return rpc.NewEmptyRpcx()
// 	}
// 	if actorID == 0 {
// 		xlog.Error("ActorCall传入ActorID不能为空")
// 		return rpc.NewEmptyRpcx()
// 	}

// 	session := as.getSession(entityConf)
// 	if session == nil {
// 		return rpc.NewEmptyRpcx()
// 	}
// 	cmd := cmdhelper.ToCmd(msg, response, actorID)
// 	if cmd == 0 {
// 		return rpc.NewEmptyRpcx()
// 	}
// 	if entityConf.AppID == gox.AppConf.Eid {
// 		if response, err := cmdhelper.CallEvt(cmd, as.context, msg); err == nil {
// 			rpcx := rpc.NewRpcx(as.context, response)
// 			defer rpcx.Run(true)
// 			return rpcx
// 		} else {
// 			xlog.Warn("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
// 			return rpc.NewEmptyRpcx()
// 		}
// 	}
// 	return session.CallByCmd(cmd, msg, response)
// }

// func (as *ActorSystem) Start() {
// 	as.EtcdComponent.OnPut = as.onPut
// 	as.EtcdComponent.OnDel = as.onDel
// 	timeoutCtx, timeoutCancelFunc := context.WithCancel(as.context)
// 	go as.checkTimeout(timeoutCtx)
// 	var err error
// 	as.es, err = etcd.NewEtcdConf(gox.AppConf.Etcd, as)
// 	timeoutCancelFunc()
// 	if err != nil {
// 		xlog.Fatal("actor启动失败 [%v]", err)
// 		return
// 	}
// 	as.es.Get(as.actorPrefix, true)
// }
// func (as *ActorSystem) Stop() {
// 	if as.es != nil {
// 		as.es.Close()
// 	}
// }
// func (as *ActorSystem) checkTimeout(ctx context.Context) {
// 	select {
// 	case <-ctx.Done():
// 		// 被取消，直接返回
// 		return
// 	case <-time.After(time.Second * 5):
// 		xlog.Fatal("请检查你的etcd服务是否有开启")
// 	}
// }

// func (as *ActorSystem) onPut(kv *mvccpb.KeyValue) {
// 	if kv.Value == nil {
// 		return
// 	}
// 	key := string(kv.Key)

// 	value, ok := as.keyToActorConf[key]
// 	if !ok {
// 		value = ActorEntity{}
// 	}
// 	if err := json.Unmarshal(kv.Value, &value); err != nil {
// 		xlog.Error("put actor err[%v]", err)
// 		return
// 	}
// 	if !ok {
// 		as.keyToActorConf[key] = value
// 	}
// 	xlog.Debug("actor注册 %v", string(kv.Value))
// }

// func (as *ActorSystem) onDel(kv *mvccpb.KeyValue) {
// 	key := string(kv.Key)
// 	delete(as.keyToActorConf, key)
// }