package network

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/util"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type (
	ServiceCrtl struct {
		nw           *NetWork
		es           *etcd.EtcdService
		lock         sync.RWMutex
		keyToService map[string]*ServiceConfig
		idToService  map[uint]*ServiceConfig
		curService   *ServiceConfig
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
		ServiceID uint
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
func (sc *ServiceConfig) GetServiceID() uint {
	return sc.ServiceID
}
func (sc *ServiceConfig) GetServiceType() string {
	return sc.ServiceType
}
func (sc *ServiceConfig) GetVersion() string {
	return sc.Version
}

func newServiceReg(nw *NetWork) *ServiceCrtl {
	return &ServiceCrtl{
		nw:           nw,
		keyToService: make(map[string]*ServiceConfig),
		idToService:  make(map[uint]*ServiceConfig),
	}
}

func convertKey(sc *ServiceConfig) string {
	key := "services/" + sc.ServiceType + "/" + util.ValToString(sc.ServiceID)
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

func (reg *ServiceCrtl) Start(ctx context.Context) {
	reg.curService = &ServiceConfig{
		ServiceID:    reg.nw.engine.ServiceID(),
		ServiceType:  reg.nw.engine.ServiceType(),
		Version:      reg.nw.engine.Version(),
		OutsideAddr:  reg.nw.GetOutsideAddr(),
		InteriorAddr: reg.nw.GetInteriorAddr(),
		RPCAddr:      reg.nw.engine.GetRPC().GetAddr(),
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(ctx)
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
func (reg *ServiceCrtl) Stop() {
	// if reg.curService != nil {
	// 	key := convertKey(reg.curService)
	// 	reg.es.Del(key)
	// }
	if reg.es != nil {
		reg.es.Close()
	}
}
func (reg *ServiceCrtl) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

//GetServiceConfByID 通过id获取服务配置
func (reg *ServiceCrtl) GetServiceConfByID(id uint) types.IServiceConfig {
	defer reg.lock.RUnlock()
	reg.lock.RLock()
	if conf, ok := reg.idToService[id]; ok {
		return conf
	}
	return nil
}

//GetServiceConfListByType 获取对应类型的所有服务配置
func (reg *ServiceCrtl) GetServiceConfListByType(serviceType string) []types.IServiceConfig {
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

func (reg *ServiceCrtl) get(resp *clientv3.GetResponse) {
	if resp == nil || resp.Kvs == nil {
		return
	}
	defer reg.lock.Unlock()
	reg.lock.Lock()
	for i := range resp.Kvs {
		reg.onPut(resp.Kvs[i])
	}
}
func (reg *ServiceCrtl) onPut(kv *mvccpb.KeyValue) {
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
func (reg *ServiceCrtl) put(kv *mvccpb.KeyValue) {
	defer reg.lock.Unlock()
	reg.lock.Lock()
	reg.onPut(kv)
}

func (reg *ServiceCrtl) del(kv *mvccpb.KeyValue) {
	reg.lock.Lock()
	defer reg.lock.Unlock()
	key := string(kv.Key)
	if service, ok := reg.keyToService[key]; ok {
		delete(reg.keyToService, key)
		delete(reg.idToService, service.ServiceID)
		xlog.Info("服务注销 sid:[%d] type:[%s]", service.ServiceID, service.ServiceType)
	}
}
