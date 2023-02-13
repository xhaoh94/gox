package network

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

type (
	ServiceSystem struct {
		etcd.EtcdComponent
		context      context.Context
		es           *etcd.EtcdConf
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

func newServiceSystem(ctx context.Context) *ServiceSystem {
	return &ServiceSystem{
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

func (ss *ServiceSystem) Start() {
	appConf := gox.AppConf
	ss.EtcdComponent.OnPut = ss.onPut
	ss.EtcdComponent.OnDel = ss.onDel
	ss.curService = ServiceEntity{
		EID:          appConf.Eid,
		EType:        appConf.EType,
		Version:      appConf.Version,
		OutsideAddr:  appConf.OutsideAddr,
		InteriorAddr: appConf.InteriorAddr,
		RpcAddr:      appConf.RpcAddr,
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(ss.context)
	go ss.checkTimeout(timeoutCtx)
	var err error
	ss.es, err = etcd.NewEtcdConf(gox.AppConf.Etcd, ss)
	timeoutCancelFunc()
	if err != nil {
		xlog.Fatal("服务注册失败 [%v]", err)
		return
	}
	key := convertKey(ss.curService)
	value := convertValue(ss.curService)
	ss.es.Put(key, value)
	ss.es.Get("services/", true)
}
func (ss *ServiceSystem) Stop() {
	// if ss.curService != nil {
	// 	key := convertKey(ss.curService)
	// 	ss.es.Del(key)
	// }
	if ss.es != nil {
		ss.es.Close()
	}
}
func (ss *ServiceSystem) checkTimeout(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 被取消，直接返回
		return
	case <-time.After(time.Second * 5):
		xlog.Fatal("请检查你的etcd服务是否有开启")
	}
}

// GetServiceEntityByID 通过id获取服务配置
func (ss *ServiceSystem) GetServiceEntityByID(id uint) types.IServiceEntity {
	defer ss.RUnlock()
	ss.RLock()
	if conf, ok := ss.idToService[id]; ok {
		return conf
	}
	return nil
}

// GetServiceEntitysByType 获取对应类型的所有服务配置
func (ss *ServiceSystem) GetServiceEntitysByType(appType string) []types.IServiceEntity {
	defer ss.RUnlock()
	ss.RLock()
	list := make([]types.IServiceEntity, 0)
	for _, v := range ss.idToService {
		// v := ss.idToService[k]
		if v.EType == appType {
			list = append(list, v)
		}
	}
	return list
}

func (ss *ServiceSystem) onPut(kv *mvccpb.KeyValue) {
	if kv.Value == nil {
		return
	}
	key := string(kv.Key)
	service, err := newServiceConfig(kv.Value)
	if err != nil {
		xlog.Error("解析服务注册配置错误[%v]", err)
		return
	}
	ss.idToService[service.EID] = service
	ss.keyToService[key] = service
	xlog.Info("服务注册发现 sid:[%d] type:[%s] version:[%s]", service.EID, service.EType, service.Version)
}
func (ss *ServiceSystem) onDel(kv *mvccpb.KeyValue) {
	key := string(kv.Key)
	if service, ok := ss.keyToService[key]; ok {
		delete(ss.keyToService, key)
		delete(ss.idToService, service.EID)
		xlog.Info("服务注销 sid:[%d] type:[%s]", service.EID, service.EType)
	}
}
