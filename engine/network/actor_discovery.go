package network

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

var ()

type (
	ActorDiscovery struct {
		engine         types.IEngine
		context        context.Context
		actorPrefix    string
		es             *etcd.EtcdService
		keyLock        sync.RWMutex
		keyToActorConf map[string]ActorEntity
	}
	ActorEntity struct {
		ActorID   uint32
		ServiceID uint
	}
)

func newActorDiscovery(engine types.IEngine, ctx context.Context) *ActorDiscovery {
	return &ActorDiscovery{
		context:        ctx,
		engine:         engine,
		actorPrefix:    "location/actor",
		keyToActorConf: make(map[string]ActorEntity),
	}
}

func (crtl *ActorDiscovery) parseFn(aid uint32, fn interface{}) uint32 {
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
		crtl.engine.NetWork().RegisterRType(cmd, in)
	}
	crtl.engine.Event().Bind(cmd, fn)
	return cmd
}

func (crtl *ActorDiscovery) Add(actor types.IActorEntity) {
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
	reg := &ActorEntity{ActorID: aid, ServiceID: crtl.engine.EID()}
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
func (crtl *ActorDiscovery) Del(actor types.IActorEntity) {
	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	aid := actor.ActorID()
	if aid == 0 {
		xlog.Error("Actor没有初始化ID")
		return
	}
	cmdList := actor.GetCmdList()
	for index := range cmdList {
		cmd := cmdList[index]
		crtl.engine.NetWork().UnRegisterRType(cmd)
		crtl.engine.Event().UnBind(cmd)
	}

	key := fmt.Sprintf(crtl.actorPrefix+"/%d", aid)
	if _, ok := crtl.keyToActorConf[key]; ok {
		crtl.es.Del(key)
	}
	actor.Destroy()
}
func (crtl *ActorDiscovery) Get(actorID uint32) (ActorEntity, bool) {
	defer crtl.keyLock.RUnlock()
	crtl.keyLock.RLock()
	key := fmt.Sprintf(crtl.actorPrefix+"/%d", actorID)
	actor, ok := crtl.keyToActorConf[key]
	if !ok {
		xlog.Error("找不到对应的Actor[%d]", actorID)
		return ActorEntity{}, false
	}
	return actor, true
}
func (crtl *ActorDiscovery) getSession(actorID uint32) types.ISession {
	conf, ok := crtl.Get(actorID)
	if !ok {
		return nil
	}
	svConf := crtl.engine.NetWork().ServiceDiscovery().GetServiceConfByID(conf.ServiceID)
	if svConf == nil {
		xlog.Error("Actor没有找到服务 ServiceID:[%s]", conf.ServiceID)
		return nil
	}
	session := crtl.engine.NetWork().GetSessionByAddr(svConf.GetInteriorAddr())
	if session == nil {
		xlog.Error("Actor没有找到session[%d]", svConf.GetInteriorAddr())
		return nil
	}
	return session
}

func (crtl *ActorDiscovery) Send(actorID uint32, msg interface{}) bool {
	session := crtl.getSession(actorID)
	if session == nil {
		return false
	}
	cmd := cmdhelper.ToCmd(msg, nil, actorID)
	return session.Send(cmd, msg)
}
func (crtl *ActorDiscovery) Call(actorID uint32, msg interface{}, response interface{}) types.IXRPC {
	session := crtl.getSession(actorID)
	if session == nil {
		dr := rpc.NewDefaultRpc(0, context.TODO(), response)
		defer dr.Run(false)
		return dr
	}
	return session.ActorCall(actorID, msg, response)
}

func (crtl *ActorDiscovery) Start() {

	timeoutCtx, timeoutCancelFunc := context.WithCancel(crtl.context)
	go crtl.checkTimeout(timeoutCtx)
	var err error
	crtl.es, err = etcd.NewEtcdService(crtl.get, crtl.put, crtl.del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("actor启动失败 [%v]", err)
		return
	}
	crtl.es.Get(crtl.actorPrefix, true)
}
func (crtl *ActorDiscovery) Stop() {
	if crtl.es != nil {
		crtl.es.Close()
	}
}
func (crtl *ActorDiscovery) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

func (crtl *ActorDiscovery) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}

	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	for i := range resp.Kvs {
		crtl.onPut(resp.Kvs[i])
	}
}
func (crtl *ActorDiscovery) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)

	value, ok := crtl.keyToActorConf[key]
	if !ok {
		value = ActorEntity{}
	}
	if err := json.Unmarshal(kv.Value, &value); err != nil {
		xlog.Error("put actor err[%v]", err)
		return
	}
	if !ok {
		crtl.keyToActorConf[key] = value
	}
	xlog.Debug("actor注册 %v", string(kv.Value))
}
func (crtl *ActorDiscovery) put(kv *mvccpb.KeyValue) {
	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	crtl.onPut(kv)
}
func (crtl *ActorDiscovery) del(kv *mvccpb.KeyValue) {
	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	key := string(kv.Key)
	delete(crtl.keyToActorConf, key)
}
