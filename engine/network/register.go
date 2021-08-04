package network

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type (
	ServiceReg struct {
		nw           *NetWork
		es           *etcd.EtcdService
		lock         sync.RWMutex
		keyToService map[string]*ServiceConfig
		idToService  map[string]*ServiceConfig
	}

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

func (sc *ServiceConfig) GetRpcAddr() string {
	return sc.RPCAddr
}
func (sc *ServiceConfig) GetOutsideAddr() string {
	return sc.OutsideAddr
}
func (sc *ServiceConfig) GetInteriorAddr() string {
	return sc.InteriorAddr
}
func (sc *ServiceConfig) GetServiceID() string {
	return sc.ServiceID
}
func (sc *ServiceConfig) GetServiceType() string {
	return sc.ServiceType
}
func (sc *ServiceConfig) GetVersion() string {
	return sc.Version
}

func newServiceReg(nw *NetWork) *ServiceReg {
	return &ServiceReg{
		nw:           nw,
		keyToService: make(map[string]*ServiceConfig),
		idToService:  make(map[string]*ServiceConfig),
	}
}

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

func (reg *ServiceReg) Start(ctx context.Context) {
	sc := &ServiceConfig{
		ServiceID:    reg.nw.engine.GetServiceID(),
		ServiceType:  reg.nw.engine.GetServiceType(),
		Version:      reg.nw.engine.GetServiceVersion(),
		OutsideAddr:  reg.nw.GetOutsideAddr(),
		InteriorAddr: reg.nw.GetInteriorAddr(),
		RPCAddr:      reg.nw.GetRpcAddr(),
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(ctx)
	go reg.checkTimeout(timeoutCtx)
	var err error
	reg.es, err = etcd.NewEtcdService(reg.get, reg.put, reg.del, ctx)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("服务注册失败 [%v]", err)
		return
	}
	key := convertKey(sc)
	value := convertValue(sc)
	reg.es.Put(key, value)
	reg.es.Get("services/", true)
}
func (reg *ServiceReg) Stop() {
	if reg.es != nil {
		reg.es.Close()
	}
}
func (reg *ServiceReg) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

//GetServiceConfByID 通过id获取服务配置
func (reg *ServiceReg) GetServiceConfByID(id string) types.IServiceConfig {
	defer reg.lock.RUnlock()
	reg.lock.RLock()
	if conf, ok := reg.idToService[id]; ok {
		return conf
	}
	return nil
}

//GetServiceConfListByType 获取对应类型的所有服务配置
func (reg *ServiceReg) GetServiceConfListByType(serviceType string) []types.IServiceConfig {
	defer reg.lock.RUnlock()
	reg.lock.RLock()
	list := make([]types.IServiceConfig, 0)
	for k := range reg.idToService {
		v := reg.idToService[k]
		if v.ServiceType == serviceType {
			list = append(list, v)
		}
	}
	return list
}

func (reg *ServiceReg) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	reg.lock.Lock()
	for i := range resp.Kvs {
		reg.onPut(resp.Kvs[i])
	}
	reg.lock.Unlock()
}
func (reg *ServiceReg) onPut(kv *mvccpb.KeyValue) {
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
	xlog.Info("服务注册 id:[%s] type:[%s] version:[%s]", service.ServiceID, service.ServiceType, service.Version)
}
func (reg *ServiceReg) put(kv *mvccpb.KeyValue) {
	reg.lock.Lock()
	reg.onPut(kv)
	reg.lock.Unlock()
}

func (reg *ServiceReg) del(kv *mvccpb.KeyValue) {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	key := string(kv.Key)
	service := reg.keyToService[key]
	if service != nil {
		delete(reg.keyToService, key)
		delete(reg.idToService, service.ServiceID)
		xlog.Info("服务注销 id:[%s] type:[%s]", service.ServiceID, service.ServiceType)
	}
}
