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

func (crtl *ActorCrtl) Start(ctx context.Context) {

	timeoutCtx, timeoutCancelFunc := context.WithCancel(ctx)
	go crtl.checkTimeout(timeoutCtx)
	es, err := etcd.NewEtcdService(crtl.get, crtl.put, crtl.del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("actor注册失败 [%v]", err)
		return
	}
	crtl.es = es
	crtl.es.Get(crtl.actorPrefix, true)
}
func (crtl *ActorCrtl) Stop() {
	if crtl.es != nil {
		crtl.es.Close()
	}
}
func (crtl *ActorCrtl) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

func (crtl *ActorCrtl) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}

	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	for i := range resp.Kvs {
		crtl.put(resp.Kvs[i])
	}
}
func (crtl *ActorCrtl) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)

	value, ok := crtl.keyToActorConf[key]
	if !ok {
		value = actorPool.Get().(*ActorConf)
	}
	if err := json.Unmarshal(kv.Value, value); err != nil {
		xlog.Error("put actor err[%v]", err)
		return
	}
	if !ok {
		crtl.keyToActorConf[key] = value
	}
}
func (crtl *ActorCrtl) put(kv *mvccpb.KeyValue) {
	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	crtl.onPut(kv)
}
func (crtl *ActorCrtl) del(kv *mvccpb.KeyValue) {
	defer crtl.keyLock.Unlock()
	crtl.keyLock.Lock()
	key := string(kv.Key)
	if actor, ok := crtl.keyToActorConf[key]; ok {
		actorPool.Put(actor)
		delete(crtl.keyToActorConf, key)
	}
}
