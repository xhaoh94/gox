package etcd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/app"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type (
	IEtcdComponent interface {
		put(*mvccpb.KeyValue)
		get(*clientv3.GetResponse)
		del(*mvccpb.KeyValue)
	}
	EtcdComponent struct {
		lock  sync.RWMutex
		OnPut func(kv *mvccpb.KeyValue)
		OnDel func(kv *mvccpb.KeyValue)
	}

	// EtcdService etcd
	EtcdService struct {
		isRun         bool
		leaseOverdue  bool
		conf          app.EtcdConf
		client        *clientv3.Client
		kv            clientv3.KV
		lease         clientv3.Lease
		leaseID       clientv3.LeaseID
		cancle        context.CancelFunc
		keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		etcdComponent IEtcdComponent
	}
)

func (component *EtcdComponent) RUnlock() {
	component.lock.RUnlock()

}
func (component *EtcdComponent) RLock() {
	component.lock.RLock()
}

func (component *EtcdComponent) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	defer component.lock.Unlock()
	component.lock.Lock()
	for i := range resp.Kvs {
		component.OnPut(resp.Kvs[i])
	}
}
func (component *EtcdComponent) put(kv *mvccpb.KeyValue) {
	defer component.lock.Unlock()
	component.lock.Lock()
	component.OnPut(kv)
}

func (component *EtcdComponent) del(kv *mvccpb.KeyValue) {
	defer component.lock.Unlock()
	component.lock.Lock()
	component.OnDel(kv)
}

// 创建etcd
func NewEtcdService(conf app.EtcdConf, component IEtcdComponent) (*EtcdService, error) {
	clientConf := clientv3.Config{
		Endpoints:   conf.EtcdList,
		DialTimeout: conf.EtcdTimeout,
	}
	client, err := clientv3.New(clientConf)
	if err != nil {
		return nil, err
	}
	kv := clientv3.NewKV(client)
	es := &EtcdService{
		isRun:         true,
		conf:          conf,
		client:        client,
		kv:            kv,
		etcdComponent: component,
	}
	if err := es.setLease(); err != nil {
		return nil, err
	}
	go es.listenLease()
	return es, nil
}

// 设置租约
func (es *EtcdService) setLease() error {
	lease := clientv3.NewLease(es.client)
	//设置租约时间
	leaseResp, err := lease.Grant(es.client.Ctx(), es.conf.EtcdLeaseTime)
	if err != nil {
		return err
	}
	//设置续租
	ctx, cancel := context.WithCancel(es.client.Ctx())
	leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		cancel()
		return err
	}
	es.lease = lease
	es.leaseID = leaseResp.ID
	es.cancle = cancel
	es.keepAliveChan = leaseRespChan
	return nil
}

// 监听 续租情况
func (es *EtcdService) listenLease() {
	for {
		leaseKeepResp := <-es.keepAliveChan
		if leaseKeepResp == nil {
			es.leaseOverdue = true
			return //失效跳出循环
		}
	}
}

// Del 删除
func (es *EtcdService) Del(key string) error {
	if !es.isRun {
		return errors.New("etcd 服务没有开启")
	}
	_, err := es.kv.Delete(es.client.Ctx(), key)
	return err
}

// Put 通过租约 注册服务
func (es *EtcdService) Put(key, val string) error {
	if !es.isRun {
		return errors.New("etcd 服务没有开启")
	}
	_, err := es.kv.Put(es.client.Ctx(), key, val, clientv3.WithLease(es.leaseID))
	return err
}

// RevokeLease 撤销租约
func (es *EtcdService) RevokeLease() error {
	es.cancle()
	time.Sleep(1 * time.Second)
	if es.leaseOverdue {
		return nil
	}
	_, err := es.lease.Revoke(es.client.Ctx(), es.leaseID)
	return err
}

// Get 获取
func (es *EtcdService) Get(prefix string, isWatcher bool) error {
	if !es.isRun {
		return errors.New("etcd 服务没有开启")
	}
	resp, err := es.client.Get(es.client.Ctx(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	es.etcdComponent.get(resp)
	if isWatcher {
		go es.watcher(prefix)
	}
	return nil
}

func (es *EtcdService) watcher(prefix string) {
	rch := es.client.Watch(es.client.Ctx(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		if es.isRun {
			for i := range wresp.Events {
				ev := wresp.Events[i]
				switch ev.Type {
				case mvccpb.PUT:
					es.etcdComponent.put(ev.Kv)
				case mvccpb.DELETE:
					es.etcdComponent.del(ev.Kv)
				}
			}
		}
	}
}

// Close 关闭
func (es *EtcdService) Close() {
	es.isRun = false
	es.RevokeLease()
	es.client.Close()
}
