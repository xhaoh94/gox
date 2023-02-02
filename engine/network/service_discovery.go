package network

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type (
	ServiceDiscovery struct {
		engine       types.IEngine
		context      context.Context
		es           *etcd.EtcdService
		lock         sync.RWMutex
		keyToService map[string]ServiceEntity
		idToService  map[uint]ServiceEntity
		curService   ServiceEntity
	}

	//ServiceEntity 服务组配置
	ServiceEntity struct {
		//EID 标记
		EID uint
		//EType 服务类型
		EType string
		//版本
		Version string
		//RpcAddr rpc服务地址
		RpcAddr string
		//AddrHost 外部服务地址
		OutsideAddr string
		//InteriorAddr 内部服务地址
		InteriorAddr string
	}
)

func (entity ServiceEntity) GetRpcAddr() string {
	return entity.RpcAddr
}
func (entity ServiceEntity) GetOutsideAddr() string {
	return entity.OutsideAddr
}
func (entity ServiceEntity) GetInteriorAddr() string {
	return entity.InteriorAddr
}
func (entity ServiceEntity) GetID() uint {
	return entity.EID
}
func (entity ServiceEntity) GetType() string {
	return entity.EType
}
func (entity ServiceEntity) GetVersion() string {
	return entity.Version
}

func newServiceDiscovery(engine types.IEngine, ctx context.Context) *ServiceDiscovery {
	return &ServiceDiscovery{
		engine:       engine,
		context:      ctx,
		keyToService: make(map[string]ServiceEntity),
		idToService:  make(map[uint]ServiceEntity),
	}
}

func convertKey(entity ServiceEntity) string {
	key := "services/" + entity.EType + "/" + strhelper.ValToString(entity.EID)
	return key
}

func convertValue(entity ServiceEntity) string {
	data, err := json.Marshal(entity)
	if err != nil {
		return ""
	}
	return string(data)
}

func newServiceConfig(val []byte) (ServiceEntity, error) {
	service := ServiceEntity{}
	if err := json.Unmarshal(val, &service); err != nil {
		return service, err
	}
	return service, nil
}

func (discovery *ServiceDiscovery) Start() {
	appConf := discovery.engine.AppConf()
	discovery.curService = ServiceEntity{
		EID:          appConf.Eid,
		EType:        appConf.EType,
		Version:      appConf.Version,
		OutsideAddr:  appConf.OutsideAddr,
		InteriorAddr: appConf.InteriorAddr,
		RpcAddr:      appConf.RpcAddr,
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(discovery.context)
	go discovery.checkTimeout(timeoutCtx)
	var err error
	discovery.es, err = etcd.NewEtcdService(discovery.engine.AppConf().Etcd, discovery.get, discovery.put, discovery.del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("服务注册失败 [%v]", err)
		return
	}
	key := convertKey(discovery.curService)
	value := convertValue(discovery.curService)
	discovery.es.Put(key, value)
	discovery.es.Get("services/", true)
}
func (discovery *ServiceDiscovery) Stop() {
	// if discovery.curService != nil {
	// 	key := convertKey(discovery.curService)
	// 	discovery.es.Del(key)
	// }
	if discovery.es != nil {
		discovery.es.Close()
	}
}
func (discovery *ServiceDiscovery) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

// GetServiceEntityByID 通过id获取服务配置
func (discovery *ServiceDiscovery) GetServiceEntityByID(id uint) types.IServiceEntity {
	defer discovery.lock.RUnlock()
	discovery.lock.RLock()
	if conf, ok := discovery.idToService[id]; ok {
		return conf
	}
	return nil
}

// GetServiceEntitysByType 获取对应类型的所有服务配置
func (discovery *ServiceDiscovery) GetServiceEntitysByType(serviceType string) []types.IServiceEntity {
	defer discovery.lock.RUnlock()
	discovery.lock.RLock()
	list := make([]types.IServiceEntity, 0)
	for k := range discovery.idToService {
		v := discovery.idToService[k]
		if v.EType == serviceType {
			list = append(list, v)
		}
	}
	return list
}

func (discovery *ServiceDiscovery) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	defer discovery.lock.Unlock()
	discovery.lock.Lock()
	for i := range resp.Kvs {
		discovery.onPut(resp.Kvs[i])
	}
}
func (discovery *ServiceDiscovery) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)
	service, err := newServiceConfig(kv.Value)
	if err != nil {
		xlog.Error("解析服务注册配置错误[%v]", err)
		return
	}
	discovery.idToService[service.EID] = service
	discovery.keyToService[key] = service
	xlog.Info("服务注册发现 sid:[%d] type:[%s] version:[%s]", service.EID, service.EType, service.Version)
}
func (discovery *ServiceDiscovery) put(kv *mvccpb.KeyValue) {
	defer discovery.lock.Unlock()
	discovery.lock.Lock()
	discovery.onPut(kv)
}

func (discovery *ServiceDiscovery) del(kv *mvccpb.KeyValue) {
	discovery.lock.Lock()
	defer discovery.lock.Unlock()
	key := string(kv.Key)
	if service, ok := discovery.keyToService[key]; ok {
		delete(discovery.keyToService, key)
		delete(discovery.idToService, service.EID)
		xlog.Info("服务注销 sid:[%d] type:[%s]", service.EID, service.EType)
	}
}
