package actor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type ActorReg struct {
	ActorID   uint32
	ServiceID string
	SessionID string
}

func (atr *Actor) Start(ctx context.Context) {

	timeoutCtx, timeoutCancelFunc := context.WithCancel(ctx)
	go atr.checkTimeout(timeoutCtx)
	var err error
	atr.actorEs, err = etcd.NewEtcdService(atr.get, atr.put, atr.del, ctx)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("actor注册失败 [%v]", err)
		return
	}
	atr.actorEs.Get(atr.actorPrefix, true)
}
func (atr *Actor) Stop() {
	if atr.actorEs != nil {
		atr.actorEs.Close()
	}
}
func (atr *Actor) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

func (atr *Actor) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}

	defer atr.actorRegLock.Unlock()
	atr.actorRegLock.Lock()
	for i := range resp.Kvs {
		atr.put(resp.Kvs[i])
	}
}
func (atr *Actor) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)

	value, ok := atr.keyToActorReg[key]
	if !ok {
		value = actorPool.Get().(*ActorReg)
	}
	if err := json.Unmarshal(kv.Value, value); err != nil {
		xlog.Error("put actor err[%v]", err)
		return
	}
	if !ok {
		atr.keyToActorReg[key] = value
	}
}
func (atr *Actor) put(kv *mvccpb.KeyValue) {
	defer atr.actorRegLock.Unlock()
	atr.actorRegLock.Lock()
	atr.onPut(kv)
}
func (atr *Actor) del(kv *mvccpb.KeyValue) {
	defer atr.actorRegLock.Unlock()
	atr.actorRegLock.Lock()
	key := string(kv.Key)
	if actor, ok := atr.keyToActorReg[key]; ok {
		actorPool.Put(actor)
		delete(atr.keyToActorReg, key)
	}
}
