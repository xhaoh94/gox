package gox

import (
	"context"
	"encoding/binary"
	"log"
	"os"
	"time"

	"github.com/xhaoh94/gox/engine/app"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xevent"
	"github.com/xhaoh94/gox/engine/xlog"
	yaml "gopkg.in/yaml.v3"
)

var (
// Engine types.IEngine
)

type (
	Engine struct {
		ctx         context.Context
		ctxCancelFn context.CancelFunc
		appConf     app.AppConf
		mainModule  types.IModule
		event       types.IEvent
		network     types.INetwork
		endian      binary.ByteOrder
	}
)

func NewEngine(appConfPath string) *Engine {
	e := new(Engine)
	e.appConf = loadConf(appConfPath)
	e.ctx, e.ctxCancelFn = context.WithCancel(context.Background())
	e.event = xevent.New()
	e.endian = binary.LittleEndian
	return e
}

func loadConf(appConfPath string) app.AppConf {
	AppCfg := app.AppConf{}
	bytes, err := os.ReadFile(appConfPath)
	if err != nil {
		log.Fatalf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return AppCfg
	}
	err = yaml.Unmarshal(bytes, &AppCfg)
	if err != nil {
		log.Fatalf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return AppCfg
	}
	AppCfg.Network.ReConnectInterval *= time.Second
	AppCfg.Network.Heartbeat *= time.Second
	AppCfg.Network.ConnectTimeout *= time.Second
	AppCfg.Network.ReadTimeout *= time.Second
	AppCfg.Etcd.EtcdTimeout *= time.Second
	return AppCfg
}

func (engine *Engine) Context() context.Context {
	return engine.ctx
}

func (engine *Engine) Endian() binary.ByteOrder {
	return engine.endian
}

func (engine *Engine) Event() types.IEvent {
	return engine.event
}

func (engine *Engine) NetWork() types.INetwork {
	return engine.network
}

func (engine *Engine) AppConf() app.AppConf {
	return engine.appConf
}

// Start 启动
func (engine *Engine) Start() {
	if engine.mainModule == nil {
		log.Fatalf("没有设置主模块")
		return
	}
	appConf := engine.appConf
	xlog.Init(appConf.Log)
	xlog.Info("服务启动[sid:%d,type:%s,ver:%s]", appConf.Eid, appConf.EType, appConf.Version)
	xlog.Info("[ByteOrder:%s]", engine.endian.String())
	engine.network.Init()
	engine.mainModule.Init(engine.mainModule, engine, func() {
		engine.network.Rpc().Serve()
	})
}

// Shutdown 关闭
func (engine *Engine) Shutdown() {
	engine.ctxCancelFn()
	engine.mainModule.Destroy(engine.mainModule)
	engine.network.Destroy()
	xlog.Info("服务退出[sid:%d]", engine.appConf.Eid)
	xlog.Destroy()
}

////////////////////////////////////////////////////////////////

// // SetOutsideService 设置外部服务类型
// func (engine *Engine) SetOutsideService(ser types.IService, codec types.ICodec) {
// 	engine.network.SetOutsideService(ser, engine.appConf.OutsideAddr, codec)
// }

// // SetInteriorService 设置内部服务类型
// func (engine *Engine) SetInteriorService(ser types.IService, codec types.ICodec) {
// 	engine.network.SetInteriorService(ser, engine.appConf.InteriorAddr, codec)
// }

// SetModule 设置网络模块
func (engine *Engine) SetNetWork(network types.INetwork) {
	engine.network = network
}

// SetEndian 设置大小端
func (engine *Engine) SetEndian(endian binary.ByteOrder) {
	engine.endian = endian
}

// SetModule 设置初始模块
func (engine *Engine) SetModule(module types.IModule) {
	engine.mainModule = module
}

////////////////////////////////////////////////////////////
