package network

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type (
	ActorSystem struct {
		context        context.Context
		actorPrefix    string
		es             *etcd.EtcdService
		keyLock        sync.RWMutex
		keyToActorConf map[string]ActorEntity
	}
	ActorEntity struct {
		ActorID uint32
		EID     uint
	}
)

func newActorSystem(ctx context.Context) *ActorSystem {
	return &ActorSystem{
		context:        ctx,
		actorPrefix:    "location/actor",
		keyToActorConf: make(map[string]ActorEntity),
	}
}

func (discovery *ActorSystem) parseFn(aid uint32, fn interface{}) uint32 {
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
		gox.NetWork.RegisterRType(cmd, in)
	}
	gox.Event.Bind(cmd, fn)
	return cmd
}

func (discovery *ActorSystem) Add(actor types.IActorEntity) {
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
	reg := &ActorEntity{ActorID: aid, EID: gox.AppConf.Eid}
	b, err := json.Marshal(reg)
	if err != nil {
		xlog.Error("Actor解析Json错误[%v]", err)
		return
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := discovery.parseFn(aid, fn); cmd != 0 {
			actor.SetCmdList(cmd)
		}
	}

	key := fmt.Sprintf(discovery.actorPrefix+"/%d", aid)
	discovery.es.Put(key, string(b))
}
func (discovery *ActorSystem) Del(actor types.IActorEntity) {
	defer discovery.keyLock.Unlock()
	discovery.keyLock.Lock()
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	cmdList := actor.GetCmdList()
	for index := range cmdList {
		cmd := cmdList[index]
		gox.NetWork.UnRegisterRType(cmd)
		gox.Event.UnBind(cmd)
	}

	key := fmt.Sprintf(discovery.actorPrefix+"/%d", aid)
	if _, ok := discovery.keyToActorConf[key]; ok {
		discovery.es.Del(key)
	}
	actor.Destroy()
}
func (discovery *ActorSystem) Get(actorID uint32) (ActorEntity, bool) {
	defer discovery.keyLock.RUnlock()
	discovery.keyLock.RLock()
	key := fmt.Sprintf(discovery.actorPrefix+"/%d", actorID)
	actor, ok := discovery.keyToActorConf[key]
	if !ok {
		xlog.Error("找不到对应的Actor[%d]", actorID)
		return ActorEntity{}, false
	}
	return actor, true
}
func (discovery *ActorSystem) getSession(actorID uint32) types.ISession {
	conf, ok := discovery.Get(actorID)
	if !ok {
		return nil
	}
	svConf := gox.ServiceSystem.GetServiceEntityByID(conf.EID)
	if svConf == nil {
		xlog.Error("Actor没有找到服务 ServiceID:[%s]", conf.EID)
		return nil
	}
	session := gox.NetWork.GetSessionByAddr(svConf.GetInteriorAddr())
	if session == nil {
		xlog.Error("Actor没有找到session[%d]", svConf.GetInteriorAddr())
		return nil
	}
	return session
}

func (discovery *ActorSystem) Send(actorID uint32, msg interface{}) bool {
	session := discovery.getSession(actorID)
	if session == nil {
		return false
	}
	cmd := cmdhelper.ToCmd(msg, nil, actorID)
	return session.Send(cmd, msg)
}
func (discovery *ActorSystem) Call(actorID uint32, msg interface{}, response interface{}) types.IRpcx {
	session := discovery.getSession(actorID)
	if session == nil {
		dr := rpc.NewDefaultRpc(0, context.TODO(), response)
		defer dr.Run(false)
		return dr
	}
	return session.ActorCall(actorID, msg, response)
}

func (discovery *ActorSystem) Start() {

	timeoutCtx, timeoutCancelFunc := context.WithCancel(discovery.context)
	go discovery.checkTimeout(timeoutCtx)
	var err error
	discovery.es, err = etcd.NewEtcdService(gox.AppConf.Etcd, discovery.get, discovery.put, discovery.del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("actor启动失败 [%v]", err)
		return
	}
	discovery.es.Get(discovery.actorPrefix, true)
}
func (discovery *ActorSystem) Stop() {
	if discovery.es != nil {
		discovery.es.Close()
	}
}
func (discovery *ActorSystem) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

func (discovery *ActorSystem) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}

	defer discovery.keyLock.Unlock()
	discovery.keyLock.Lock()
	for i := range resp.Kvs {
		discovery.onPut(resp.Kvs[i])
	}
}
func (discovery *ActorSystem) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)

	value, ok := discovery.keyToActorConf[key]
	if !ok {
		value = ActorEntity{}
	}
	if err := json.Unmarshal(kv.Value, &value); err != nil {
		xlog.Error("put actor err[%v]", err)
		return
	}
	if !ok {
		discovery.keyToActorConf[key] = value
	}
	xlog.Debug("actor注册 %v", string(kv.Value))
}
func (discovery *ActorSystem) put(kv *mvccpb.KeyValue) {
	defer discovery.keyLock.Unlock()
	discovery.keyLock.Lock()
	discovery.onPut(kv)
}
func (discovery *ActorSystem) del(kv *mvccpb.KeyValue) {
	defer discovery.keyLock.Unlock()
	discovery.keyLock.Lock()
	key := string(kv.Key)
	delete(discovery.keyToActorConf, key)
}
