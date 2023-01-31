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
		//RPCAddr grpc服务地址
		RPCAddr string
		//AddrHost 外部服务地址
		OutsideAddr string
		//InteriorAddr 内部服务地址
		InteriorAddr string
		//ServiceID 标记
		ServiceID uint
		//ServiceType 服务类型
		ServiceType string
		//版本
		Version string
	}
)

func (sc ServiceEntity) GetRpcAddr() string {
	return sc.RPCAddr
}
func (sc ServiceEntity) GetOutsideAddr() string {
	return sc.OutsideAddr
}
func (sc ServiceEntity) GetInteriorAddr() string {
	return sc.InteriorAddr
}
func (sc ServiceEntity) GetServiceID() uint {
	return sc.ServiceID
}
func (sc ServiceEntity) GetServiceType() string {
	return sc.ServiceType
}
func (sc ServiceEntity) GetVersion() string {
	return sc.Version
}

func newServiceDiscovery(engine types.IEngine, ctx context.Context) *ServiceDiscovery {
	return &ServiceDiscovery{
		engine:       engine,
		context:      ctx,
		keyToService: make(map[string]ServiceEntity),
		idToService:  make(map[uint]ServiceEntity),
	}
}

func convertKey(sc ServiceEntity) string {
	key := "services/" + sc.ServiceType + "/" + strhelper.ValToString(sc.ServiceID)
	return key
}

func convertValue(sc ServiceEntity) string {
	data, err := json.Marshal(sc)
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

func (reg *ServiceDiscovery) Start() {
	reg.curService = ServiceEntity{
		ServiceID:    reg.engine.EID(),
		ServiceType:  reg.engine.EType(),
		Version:      reg.engine.Version(),
		OutsideAddr:  reg.engine.NetWork().OutsideAddr(),
		InteriorAddr: reg.engine.NetWork().InteriorAddr(),
		RPCAddr:      reg.engine.NetWork().Rpc().GetAddr(),
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(reg.context)
	go reg.checkTimeout(timeoutCtx)
	var err error
	reg.es, err = etcd.NewEtcdService(reg.get, reg.put, reg.del)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("服务注册失败 [%v]", err)
		return
	}
	key := convertKey(reg.curService)
	value := convertValue(reg.curService)
	reg.es.Put(key, value)
	reg.es.Get("services/", true)
}
func (reg *ServiceDiscovery) Stop() {
	// if reg.curService != nil {
	// 	key := convertKey(reg.curService)
	// 	reg.es.Del(key)
	// }
	if reg.es != nil {
		reg.es.Close()
	}
}
func (reg *ServiceDiscovery) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

// GetServiceConfByID 通过id获取服务配置
func (reg *ServiceDiscovery) GetServiceConfByID(id uint) types.IServiceEntity {
	defer reg.lock.RUnlock()
	reg.lock.RLock()
	if conf, ok := reg.idToService[id]; ok {
		return conf
	}
	return nil
}

// GetServiceConfListByType 获取对应类型的所有服务配置
func (reg *ServiceDiscovery) GetServiceConfListByType(serviceType string) []types.IServiceEntity {
	defer reg.lock.RUnlock()
	reg.lock.RLock()
	list := make([]types.IServiceEntity, 0)
	for k := range reg.idToService {
		v := reg.idToService[k]
		if v.ServiceType == serviceType {
			list = append(list, v)
		}
	}
	return list
}

func (reg *ServiceDiscovery) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	defer reg.lock.Unlock()
	reg.lock.Lock()
	for i := range resp.Kvs {
		reg.onPut(resp.Kvs[i])
	}
}
func (reg *ServiceDiscovery) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)
	service, err := newServiceConfig(kv.Value)
	if err != nil {
		xlog.Error("解析服务注册配置错误[%v]", err)
		return
	}
	reg.idToService[service.ServiceID] = service
	reg.keyToService[key] = service
	xlog.Info("服务注册发现 sid:[%d] type:[%s] version:[%s]", service.ServiceID, service.ServiceType, service.Version)
}
func (reg *ServiceDiscovery) put(kv *mvccpb.KeyValue) {
	defer reg.lock.Unlock()
	reg.lock.Lock()
	reg.onPut(kv)
}

func (reg *ServiceDiscovery) del(kv *mvccpb.KeyValue) {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	key := string(kv.Key)
	if service, ok := reg.keyToService[key]; ok {
		delete(reg.keyToService, key)
		delete(reg.idToService, service.ServiceID)
		xlog.Info("服务注销 sid:[%d] type:[%s]", service.ServiceID, service.ServiceType)
	}
}
