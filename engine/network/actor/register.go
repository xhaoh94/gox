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

func (atr *ActorCrtl) Start(ctx context.Context) {

	timeoutCtx, timeoutCancelFunc := context.WithCancel(ctx)
	go atr.checkTimeout(timeoutCtx)
	var err error
	atr.actorEs, err = etcd.NewEtcdService(atr.get, atr.put, atr.del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("actor注册失败 [%v]", err)
		return
	}
	atr.actorEs.Get(atr.actorPrefix, true)
}
func (atr *ActorCrtl) Stop() {
	if atr.actorEs != nil {
		atr.actorEs.Close()
	}
}
func (atr *ActorCrtl) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

func (atr *ActorCrtl) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}

	defer atr.keyLock.Unlock()
	atr.keyLock.Lock()
	for i := range resp.Kvs {
		atr.put(resp.Kvs[i])
	}
}
func (atr *ActorCrtl) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)

	value, ok := atr.keyToActorConf[key]
	if !ok {
		value = actorPool.Get().(*ActorConf)
	}
	if err := json.Unmarshal(kv.Value, value); err != nil {
		xlog.Error("put actor err[%v]", err)
		return
	}
	if !ok {
		atr.keyToActorConf[key] = value
	}
}
func (atr *ActorCrtl) put(kv *mvccpb.KeyValue) {
	defer atr.keyLock.Unlock()
	atr.keyLock.Lock()
	atr.onPut(kv)
}
func (atr *ActorCrtl) del(kv *mvccpb.KeyValue) {
	defer atr.keyLock.Unlock()
	atr.keyLock.Lock()
	key := string(kv.Key)
	if actor, ok := atr.keyToActorConf[key]; ok {
		actorPool.Put(actor)
		delete(atr.keyToActorConf, key)
	}
}
