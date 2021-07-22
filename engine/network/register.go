package network

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

var (
	es           *etcd.EtcdService
	curService   *ServiceConfig
	lock         sync.RWMutex
	keyToService map[string]*ServiceConfig
	idToService  map[string]*ServiceConfig
)

type (
	//ServiceConfig 服务组配置
	ServiceConfig struct {
		//RPCAddr grpc服务地址
		RPCAddr string
		//AddrHost 外部服务地址
		OutsideAddr string
		//InteriorAddr 内部服务地址
		InteriorAddr string
		//ServiceID 标记
		ServiceID string
		//ServiceType 服务类型
		ServiceType string
		//版本
		Version string
	}
)

func convertKey(sc *ServiceConfig) string {
	key := "services/" + string(sc.ServiceType) + "/" + sc.ServiceID
	return key
}

func convertValue(sc *ServiceConfig) string {
	data, err := json.Marshal(sc)
	if err != nil {
		return ""
	}
	return string(data)
}

func newServiceConfig(val []byte) (*ServiceConfig, error) {
	service := &ServiceConfig{}
	if err := json.Unmarshal(val, service); err != nil {
		return nil, err
	}
	return service, nil
}

func registerService(outsideAddr string, interiorAddr string, rpcAddr string) {
	keyToService = make(map[string]*ServiceConfig)
	idToService = make(map[string]*ServiceConfig)

	curService = &ServiceConfig{
		ServiceID:    app.SID,
		ServiceType:  app.ServiceType,
		OutsideAddr:  outsideAddr,
		InteriorAddr: interiorAddr,
		RPCAddr:      rpcAddr,
		Version:      app.Version,
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(context.TODO())
	go checkTimeout(timeoutCtx)
	var err error
	es, err = etcd.NewEtcdService(get, put, del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("new etcd service err [%v]", err)
		return
	}
	key := convertKey(curService)
	value := convertValue(curService)
	es.Put(key, value)
	es.Get("services/", true)
}
func unRegisterService() {

	if es != nil {
		es.Close()
	}
}
func checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Error("check if the etcd service is open")
		xlog.Fatal("exit")
	}
}

//CurServiceConf 获取当前服务
func CurServiceConf() *ServiceConfig {
	return curService
}

//GetServiceConfByID 通过id获取服务配置
func GetServiceConfByID(id string) *ServiceConfig {
	defer lock.RUnlock()
	lock.RLock()
	if conf, ok := idToService[id]; ok {
		return conf
	}
	return nil
}

//GetServiceConfListByType 获取对应类型的所有服务配置
func GetServiceConfListByType(serviceType string) []*ServiceConfig {
	defer lock.RUnlock()
	lock.RLock()
	list := make([]*ServiceConfig, 0)
	for k := range idToService {
		v := idToService[k]
		if v.ServiceType == serviceType {
			list = append(list, v)
		}
	}
	return list
}

func get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	lock.Lock()
	for i := range resp.Kvs {
		onPut(resp.Kvs[i])
	}
	lock.Unlock()
}
func onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)
	service, err := newServiceConfig(kv.Value)
	if err != nil {
		xlog.Error("register service err [%v]", err)
		return
	}
	idToService[service.ServiceID] = service
	keyToService[key] = service
	xlog.Info("register service for id:[%s] type:[%s] version:[%s]", service.ServiceID, service.ServiceType, service.Version)
}
func put(kv *mvccpb.KeyValue) {
	lock.Lock()
	onPut(kv)
	lock.Unlock()
}

func del(kv *mvccpb.KeyValue) {
	lock.Lock()
	defer lock.Unlock()
	key := string(kv.Key)
	service := keyToService[key]
	if service != nil {
		delete(keyToService, key)
		delete(idToService, service.ServiceID)
		xlog.Info("unRegister service for id:[%s] type:[%s]", service.ServiceID, service.ServiceType)
	}
}
