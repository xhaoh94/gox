package etcd

import (
	"context"
	"errors"

	"github.com/xhaoh94/gox/app"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type (
	putFn func(kv *mvccpb.KeyValue)
	getFn func(resp *clientv3.GetResponse)
	delFn func(kv *mvccpb.KeyValue)
)

//EtcdService etcd
type EtcdService struct {
	isRun         bool
	client        *clientv3.Client
	kv            clientv3.KV
	lease         clientv3.Lease
	leaseID       clientv3.LeaseID
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	putFn         putFn
	getFn         getFn
	delFn         delFn
}

func GetEtcdConf() clientv3.Config {
	return clientv3.Config{Endpoints: app.GetAppCfg().Etcd.EtcdList, DialTimeout: app.GetAppCfg().Etcd.EtcdTimeout}
}

//NewEtcdService 创建etcd
func NewEtcdService(get getFn, put putFn, del delFn, ctx context.Context) (*EtcdService, error) {
	conf := clientv3.Config{
		Endpoints:   app.GetAppCfg().Etcd.EtcdList,
		DialTimeout: app.GetAppCfg().Etcd.EtcdTimeout,
		Context:     ctx,
	}
	client, err := clientv3.New(conf)
	if err != nil {
		return nil, err
	}
	kv := clientv3.NewKV(client)
	es := &EtcdService{
		isRun:  true,
		client: client,
		kv:     kv,
		putFn:  put,
		getFn:  get,
		delFn:  del,
	}
	if err := es.setLease(); err != nil {
		return nil, err
	}
	return es, nil
}

//设置租约
func (es *EtcdService) setLease() error {
	lease := clientv3.NewLease(es.client)
	//设置租约时间
	leaseResp, err := lease.Grant(es.client.Ctx(), app.GetAppCfg().Etcd.EtcdLeaseTime)
	if err != nil {
		return err
	}
	//设置续租
	ctx, cancelFunc := context.WithCancel(es.client.Ctx())
	leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)

	if err != nil {
		cancelFunc()
		return err
	}
	go es.listenLease()
	es.lease = lease
	es.leaseID = leaseResp.ID
	es.canclefunc = cancelFunc
	es.keepAliveChan = leaseRespChan
	return nil
}

//监听 续租情况
func (es *EtcdService) listenLease() {
	for {
		select {
		case leaseKeepResp := <-es.keepAliveChan:
			if leaseKeepResp == nil {
				goto END //失效跳出循环
			}
		}
	}
END:
}

//Del 删除
func (es *EtcdService) Del(key string) error {
	if !es.isRun {
		return errors.New("etcd service is close")
	}
	_, err := es.kv.Delete(es.client.Ctx(), key)
	return err
}

//Put 通过租约 注册服务
func (es *EtcdService) Put(key, val string) error {
	if !es.isRun {
		return errors.New("etcd service is close")
	}
	_, err := es.kv.Put(es.client.Ctx(), key, val, clientv3.WithLease(es.leaseID))
	return err
}

//RevokeLease 撤销租约
func (es *EtcdService) RevokeLease() error {
	es.canclefunc()
	_, err := es.lease.Revoke(es.client.Ctx(), es.leaseID)
	return err
}

//Get 获取
func (es *EtcdService) Get(prefix string, isWatcher bool) error {
	if !es.isRun {
		return errors.New("etcd service is close")
	}
	resp, err := es.client.Get(es.client.Ctx(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if es.getFn != nil {
		es.getFn(resp)
	}
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
					if es.putFn != nil {
						es.putFn(ev.Kv)
					}
				case mvccpb.DELETE:
					if es.delFn != nil {
						es.delFn(ev.Kv)
					}
				}
			}
		}
	}
}

//Close 关闭
func (es *EtcdService) Close() {
	es.isRun = false
	es.RevokeLease()
	es.client.Close()
}
