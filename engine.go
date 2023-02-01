package gox

import (
	"context"
	"encoding/binary"
	"log"

	"github.com/xhaoh94/gox/engine/app"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xevent"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	Engine struct {
		context    context.Context
		contextFn  context.CancelFunc
		appConf    app.AppConf
		mainModule types.IModule
		event      types.IEvent
		network    *network.NetWork
		endian     binary.ByteOrder
	}
)

func NewEngine(appConfPath string) *Engine {
	e := new(Engine)
	e.appConf = app.LoadAppConfig(appConfPath)
	e.event = xevent.New()
	e.context, e.contextFn = context.WithCancel(context.Background())
	e.network = network.New(e, e.context)
	e.endian = binary.LittleEndian
	return e
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
	xlog.Init(engine.AppConf().Log)
	xlog.Info("服务启动[sid:%d,type:%s,ver:%s]", engine.AppConf().Eid, engine.AppConf().EType, engine.AppConf().Version)
	xlog.Info("[ByteOrder:%s]", engine.endian.String())
	engine.network.Init()
	engine.mainModule.Init(engine.mainModule, engine, func() {
		engine.network.Serve()
	})
}

// Shutdown 关闭
func (engine *Engine) Shutdown() {
	engine.contextFn()
	engine.mainModule.Destroy(engine.mainModule)
	engine.network.Destroy()
	xlog.Info("服务退出[sid:%d]", engine.AppConf().Eid)
	xlog.Destroy()
}

////////////////////////////////////////////////////////////////

// SetOutsideService 设置外部服务类型
func (engine *Engine) SetOutsideService(ser types.IService, codec types.ICodec) {
	engine.network.SetOutsideService(ser, engine.appConf.OutsideAddr, codec)
}

// SetInteriorService 设置内部服务类型
func (engine *Engine) SetInteriorService(ser types.IService, codec types.ICodec) {
	engine.network.SetInteriorService(ser, engine.appConf.InteriorAddr, codec)
}

// SetEndian 设置大小端
func (engine *Engine) SetEndian(bo binary.ByteOrder) {
	engine.endian = bo
}

// SetModule 设置初始模块
func (engine *Engine) SetModule(m types.IModule) {
	engine.mainModule = m
}

////////////////////////////////////////////////////////////
