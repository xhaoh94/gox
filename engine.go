package gox

import (
	"context"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
)

type (
	IEngine interface {
		Start(appConfPath string)
		Shutdown()
		SetOutsideService(ser types.IService, addr string)
		SetInteriorService(ser types.IService, addr string)
		SetGrpcAddr(addr string)
		SetCodec(c types.ICodec)
	}
	Engine struct {
		sid     string
		stype   string
		version string

		context   context.Context
		contextFn context.CancelFunc
		mol       types.IModule
		codec     types.ICodec
		event     types.IEvent
		nw        *network.NetWork
	}
)

func NewEngine(sid string, sType string, version string, m types.IModule) IEngine {
	e := new(Engine)
	e.sid = sid
	e.stype = sType
	e.version = version
	e.event = NewEvent()
	e.context, e.contextFn = context.WithCancel(context.Background())
	e.nw = network.New(e, e.context)
	e.mol = m
	return e
}

func (engine *Engine) GetCodec() types.ICodec {
	return engine.codec
}

func (engine *Engine) GetEvent() types.IEvent {
	return engine.event
}

func (engine *Engine) GetNetWork() types.INetwork {
	return engine.nw
}

func (engine *Engine) GetServiceID() string {
	return engine.sid
}
func (engine *Engine) GetServiceType() string {
	return engine.stype
}
func (engine *Engine) GetServiceVersion() string {
	return engine.version
}

//Start 启动
func (engine *Engine) Start(appConfPath string) {
	app.LoadAppConfig(appConfPath)
	xlog.Init()
	xlog.Info("服务启动[%s]", engine.sid)
	xlog.Info("服务类型[%s]", engine.stype)
	xlog.Info("服务版本[%s]", engine.version)
	engine.nw.Start()
	engine.mol.Start(engine.mol, engine)
	engine.event.Call("_start_engine_ok_")
	xlog.Info("服务启动完毕")
}

//Shutdown 关闭
func (engine *Engine) Shutdown() {
	engine.contextFn()
	engine.mol.Destroy(engine.mol)
	engine.nw.Stop()
	xlog.Info("服务退出[%s]", engine.sid)
	xlog.Destroy()
}

////////////////////////////////////////////////////////////////

//SetOutsideService 设置外部服务类型
func (engine *Engine) SetOutsideService(ser types.IService, addr string) {
	engine.nw.SetOutsideService(ser, addr)
}

//SetInteriorService 设置内部服务类型
func (engine *Engine) SetInteriorService(ser types.IService, addr string) {
	engine.nw.SetInteriorService(ser, addr)
}

//SetGrpcAddr 设置grpc服务
func (engine *Engine) SetGrpcAddr(addr string) {
	engine.nw.SetGrpcAddr(addr)
}

//SetCodec 设置解码器
func (engine *Engine) SetCodec(c types.ICodec) {
	engine.codec = c
}

// //SetModule 设置初始模块
// func (engine *Engine) SetModule(m types.IModule) {
// 	engine.mol = m
// }

////////////////////////////////////////////////////////////
