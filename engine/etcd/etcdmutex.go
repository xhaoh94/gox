package etcd

import (
	"context"
	"errors"

	"go.etcd.io/etcd/clientv3"
)

type EtcdMutex struct {
	Ttl         int64              //租约时间
	Conf        clientv3.Config    //etcd集群配置
	Key         string             //etcd的key
	ctxCancelFn context.CancelFunc //关闭续租的func
	lease       clientv3.Lease
	leaseID     clientv3.LeaseID
	txn         clientv3.Txn
}

func NewEtcdMutex(key string, tl int64, conf clientv3.Config) *EtcdMutex {
	return &EtcdMutex{
		Conf: conf,
		Ttl:  tl,
		Key:  key,
	}
}
func (em *EtcdMutex) init() error {
	var err error
	var ctx context.Context
	client, err := clientv3.New(em.Conf)
	if err != nil {
		return err
	}
	em.txn = clientv3.NewKV(client).Txn(context.TODO())
	em.lease = clientv3.NewLease(client)
	leaseResp, err := em.lease.Grant(context.TODO(), em.Ttl)
	if err != nil {
		return err
	}
	ctx, em.ctxCancelFn = context.WithCancel(context.TODO())
	em.leaseID = leaseResp.ID
	_, err = em.lease.KeepAlive(ctx, em.leaseID)
	if err != nil {
		em.ctxCancelFn()
	}
	return err
}
func (em *EtcdMutex) Lock() error {
	err := em.init()
	if err != nil {
		return err
	}
	//LOCK:
	em.txn.If(clientv3.Compare(clientv3.CreateRevision(em.Key), "=", 0)).
		Then(clientv3.OpPut(em.Key, "", clientv3.WithLease(em.leaseID))).
		Else()
	txnResp, err := em.txn.Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded { //判断txn.if条件是否成立
		return errors.New("etcd 锁失败")
	}
	return nil
}
func (em *EtcdMutex) UnLock() {
	em.ctxCancelFn()
	em.lease.Revoke(context.TODO(), em.leaseID)
}
