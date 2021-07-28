package gox

import (
	"github.com/xhaoh94/gox/engine/conf"
	"github.com/xhaoh94/gox/engine/event"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	IEngine interface {
		Start(appConfPath string)
		Shutdown()
		SetOutsideService(ser types.IService, addr string)
		SetInteriorService(ser types.IService, addr string)
		SetGrpcAddr(addr string)
		SetCodec(c types.ICodec)
		SetModule(m types.IModule)
	}
	Engine struct {
		sid     string
		stype   string
		version string

		mol   types.IModule
		codec types.ICodec
		event types.IEvent
		nw    *network.NetWork
	}
)

func NewEngine(sid string, sType string, version string) IEngine {
	e := new(Engine)
	e.sid = sid
	e.stype = sType
	e.version = version
	e.nw = network.New(e)
	e.event = event.New()
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
	conf.LoadAppConfig(appConfPath)
	xlog.Init()
	xlog.Info("server start. sid -> [%s]", engine.sid)
	xlog.Info("server type -> [%s]", engine.stype)
	xlog.Info("server version -> [%s]", engine.version)
	engine.nw.Start()
	engine.mol.Start(engine.mol, engine)
	engine.event.Call("_start_engine_ok_")
}

//Shutdown 关闭
func (engine *Engine) Shutdown() {
	engine.mol.Destroy(engine.mol)
	engine.nw.Stop()
	xlog.Info("server exit. sid -> [%s]", engine.sid)
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

//SetModule 设置初始模块
func (engine *Engine) SetModule(m types.IModule) {
	engine.mol = m
}

////////////////////////////////////////////////////////////
