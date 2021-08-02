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

type actorReg struct {
	actorID   uint32
	serviceID string
	sessionID string
}

func (atr *Actor) Start(ctx context.Context) {

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

func newActorReg(val []byte) (*actorReg, error) {
	actor := actorPool.Get().(*actorReg)
	if err := json.Unmarshal(val, actor); err != nil {
		return nil, err
	}
	return actor, nil
}

func (atr *Actor) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	for i := range resp.Kvs {
		atr.put(resp.Kvs[i])
	}
}
func (atr *Actor) put(kv *mvccpb.KeyValue) {
	atr.actorRegLock.Lock()
	defer atr.actorRegLock.Unlock()
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)
	value, err := newActorReg(kv.Value)
	if err != nil {
		xlog.Error("put actor err[%v]", err)
		return
	}
	atr.keyToActorReg[key] = value
}
func (atr *Actor) del(kv *mvccpb.KeyValue) {
	atr.actorRegLock.Lock()
	defer atr.actorRegLock.Unlock()
	key := string(kv.Key)
	if actor, ok := atr.keyToActorReg[key]; ok {
		actorPool.Put((actor))
		delete(atr.keyToActorReg, key)
	}
}
