package gox

import (
	"context"
	"encoding/binary"
	"log"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/rpc"
	"github.com/xhaoh94/gox/engine/xevent"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/xdef"
)

type (
	Engine struct {
		eid     uint
		etype   string
		version string

		context   context.Context
		contextFn context.CancelFunc
		mol       types.IModule
		event     types.IEvent
		nw        *network.NetWork
		rpc       *rpc.RPC
		endian    binary.ByteOrder
	}
)

func NewEngine(sid uint, sType string, version string) *Engine {
	e := new(Engine)
	e.eid = sid
	e.etype = sType
	e.version = version
	e.event = xevent.New()
	e.context, e.contextFn = context.WithCancel(context.Background())
	e.nw = network.New(e, e.context)
	e.rpc = rpc.New(e)
	e.endian = binary.LittleEndian
	return e
}

func (engine *Engine) GetEndian() binary.ByteOrder {
	return engine.endian
}

func (engine *Engine) Event() types.IEvent {
	return engine.event
}

func (engine *Engine) GetRPC() types.IGRPC {
	return engine.rpc
}

func (engine *Engine) GetNetWork() types.INetwork {
	return engine.nw
}

func (engine *Engine) EID() uint {
	return engine.eid
}
func (engine *Engine) EType() string {
	return engine.etype
}
func (engine *Engine) Version() string {
	return engine.version
}

// Start 启动
func (engine *Engine) Start() {
	if engine.mol == nil {
		log.Fatalf("没有设置主模块")
		return
	}
	if !app.IsLoadAppCfg() {
		xlog.Warn("没有传入ini配置,使用默认配置")
	}
	xlog.Init(engine.eid)
	xlog.Info("服务启动[sid:%d,type:%s,ver:%s]", engine.eid, engine.etype, engine.version)
	xlog.Info("[ByteOrder:%s]", engine.endian.String())
	engine.nw.Init()
	engine.rpc.Init()
	engine.mol.Init(engine.mol, engine)

	engine.rpc.Serve()
	engine.event.Run(xdef.START_ENGINE_OK)
}

// Shutdown 关闭
func (engine *Engine) Shutdown() {
	engine.contextFn()
	engine.mol.Destroy(engine.mol)
	engine.rpc.Destroy()
	engine.nw.Destroy()
	xlog.Info("服务退出[sid:%d]", engine.eid)
	xlog.Destroy()
}

////////////////////////////////////////////////////////////////

// SetOutsideService 设置外部服务类型
func (engine *Engine) SetOutsideService(ser types.IService, addr string, c types.ICodec) {
	engine.nw.SetOutsideService(ser, addr, c)
}

// SetInteriorService 设置内部服务类型
func (engine *Engine) SetInteriorService(ser types.IService, addr string, c types.ICodec) {
	engine.nw.SetInteriorService(ser, addr, c)
}

// SetGrpcAddr 设置grpc服务
func (engine *Engine) SetGrpcAddr(addr string) {
	engine.rpc.SetAddr(addr)
}

// SetEndian 设置大小端
func (engine *Engine) SetEndian(bo binary.ByteOrder) {
	engine.endian = bo
}

// SetModule 设置初始模块
func (engine *Engine) SetModule(m types.IModule) {
	engine.mol = m
}

////////////////////////////////////////////////////////////
