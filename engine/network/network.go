package network

import (
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/network/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

var (
	isRun    bool
	outside  types.IService
	interior types.IService
)

//SetOutsideService 设置外部服务
func SetOutsideService(ser types.IService, addr string) {
	outside = ser
	outside.Init(addr)
}

//SetInteriorService 设置内部服务
func SetInteriorService(ser types.IService, addr string) {
	interior = ser
	interior.Init(addr)
}

//SetGrpcAddr 设置grpc服务
func SetGrpcAddr(addr string) {
	rpc.Init(addr)
}

//Start 网络初始化入口
func Start() {
	if isRun {
		return
	}
	var outsideAddr, interiorAddr, rpcAddr string
	if interior == nil {
		xlog.Fatal("service is nil")
		return
	}
	interiorAddr = interior.GetAddr()
	rpcAddr = rpc.GetAddr()

	isRun = true
	if outside != nil {
		outsideAddr = outside.GetAddr()
		outside.Start()
	}
	interior.Start()
	rpc.Start()

	registerService(outsideAddr, interiorAddr, rpcAddr)
}

//Stop 销毁
func Stop() {
	if !isRun {
		return
	}

	isRun = false
	unRegisterService()
	rpc.Stop()

	if outside != nil {
		outside.Stop()
	}
	interior.Stop()
}

//GetSession 通过id获取Session
func GetSessionById(sid string) types.ISession {
	session := interior.GetSessionById(sid)
	if session == nil && outside != nil {
		session = outside.GetSessionById(sid)
	}
	return session
}

//GetSessionByAddr 通过地址获取Session
func GetSessionByAddr(addr string) types.ISession {
	return interior.GetSessionByAddr(addr)
}
