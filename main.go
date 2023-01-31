package gox

import (
	"context"
	"encoding/binary"
	"log"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/network"
	"github.com/xhaoh94/gox/types"
	"github.com/xhaoh94/gox/xlog"
)

var (
	Sid     uint
	Stype   string
	Version string
	NetWork types.INetwork
	Endian  binary.ByteOrder

	ctx         context.Context
	ctxCancelFn context.CancelFunc
	mainModule  types.IModule
)

func Start(sid uint, sType string, version string) {
	if mainModule == nil {
		log.Fatalf("没有设置主模块")
		return
	}

	Sid = sid
	Stype = sType
	Version = version
	if !app.IsLoadAppCfg() {
		xlog.Warn("没有传入ini配置,使用默认配置")
	}
	xlog.Init(Sid)
	xlog.Info("服务启动[sid:%d,type:%s,ver:%s]", Sid, Stype, Version)

	ctx, ctxCancelFn = context.WithCancel(context.Background())
	NetWork = network.New(ctx, sid, sType, version)
	xlog.Info("[ByteOrder:%s]", Endian.String())
	network := NetWork.(*network.NetWork)
	network.Init()
	mainModule.Init(mainModule, func() {
		network.Serve()
	})
}

// Shutdown 关闭
func Shutdown() {
	ctxCancelFn()
	NetWork.(*network.NetWork).Destroy()
	xlog.Info("服务退出[sid:%d]", Sid)
	xlog.Destroy()
	mainModule.Destroy(mainModule)

}

////////////////////////////////////////////////////////////////

// SetOutsideService 设置外部服务类型
func SetOutsideService(service types.IService, addr string, codec types.ICodec) {
	NetWork.(*network.NetWork).SetOutsideService(service, addr, codec)
}

// SetInteriorService 设置内部服务类型
func SetInteriorService(service types.IService, addr string, codec types.ICodec) {
	NetWork.(*network.NetWork).SetInteriorService(service, addr, codec)
}

// SetGrpcAddr 设置grpc服务
func SetGrpcAddr(addr string) {
	NetWork.(*network.NetWork).SetGrpcAddr(addr)
}

// SetEndian 设置大小端
func SetEndian(endian binary.ByteOrder) {
	Endian = endian
}

// SetMainModule 设置初始模块
func SetMainModule(m types.IModule) {
	mainModule = m
}

////////////////////////////////////////////////////////////
