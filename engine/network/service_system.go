package network

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/types"

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
		//标记
		AppID uint
		//服务类型
		AppType string
		//区
		Zones string
		//版本
		Version string
		//是否开启定位
		Location bool
		//rpc服务地址
		RpcAddr string
		//外部服务地址
		OutsideAddr string
		//内部服务地址
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
	return entity.AppID
}
func (entity ServiceEntity) IsLocation() bool {
	return entity.Location
}

func (entity ServiceEntity) GetType() string {
	return entity.AppType
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
	key := "services/" + entity.AppType + "/" + strhelper.ValToString(entity.AppID)
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
	appConf := gox.Config
	if len(appConf.Etcd.EtcdList) == 0 {
		logger.Error().Msg("EtcdList 为空，无法启动服务注册")
		return
	}
	ss.EtcdComponent.OnPut = ss.onPut
	ss.EtcdComponent.OnDel = ss.onDel
	ss.curService = ServiceEntity{
		AppID:        appConf.AppID,
		AppType:      appConf.AppType,
		Version:      appConf.Version,
		Location:     appConf.Location,
		OutsideAddr:  appConf.OutsideAddr,
		InteriorAddr: appConf.InteriorAddr,
		RpcAddr:      appConf.RpcAddr,
	}
	timeoutCtx, timeoutCancelFunc := context.WithCancel(ss.context)
	go ss.checkTimeout(timeoutCtx)
	var err error
	ss.es, err = etcd.NewEtcdConf(ss)
	timeoutCancelFunc()
	if err != nil {
		logger.Fatal().Err(err).Msg("EtcdList 服务注册失败")
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
		logger.Fatal().Msg("请检查你的etcd服务是否有开启")
	}
}

// 通过id获取服务配置
func (ss *ServiceSystem) GetServiceEntityByID(id uint) types.IServiceEntity {
	defer ss.RUnlock()
	ss.RLock()
	if conf, ok := ss.idToService[id]; ok {
		return conf
	}
	return nil
}
func (ss *ServiceSystem) checkOpt(entity types.IServiceEntity, opts ...types.ServiceOptionFunc) bool {
	if len(opts) > 0 {
		for _, fun := range opts {
			if !fun(entity) {
				return false
			}
		}
	}
	return true
}

// 获取对应类型的所有服务配置
func (ss *ServiceSystem) GetServiceEntitys(opts ...types.ServiceOptionFunc) []types.IServiceEntity {
	defer ss.RUnlock()
	ss.RLock()
	list := make([]types.IServiceEntity, 0)
	for _, v := range ss.idToService {
		if ss.checkOpt(v, opts...) {
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
		logger.Error().Err(err).Msg("解析服务注册配置错误")
		return
	}
	ss.idToService[service.AppID] = service
	ss.keyToService[key] = service
	logger.Info().Uint("AppID", service.AppID).Str("Type", service.AppType).Str("Version", service.Version).Msg("服务注册")
}
func (ss *ServiceSystem) onDel(kv *mvccpb.KeyValue) {
	key := string(kv.Key)
	if service, ok := ss.keyToService[key]; ok {
		delete(ss.keyToService, key)
		delete(ss.idToService, service.AppID)
		gox.Location.(*location.LocationSystem).ServiceClose(service.AppID)
		logger.Info().Uint("AppID", service.AppID).Str("Type", service.AppType).Str("Version", service.Version).Msg("服务注销")
	}
}
