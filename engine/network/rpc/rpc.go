package rpc

import (
	"net"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"

	"google.golang.org/grpc"
)

type (
	RPC struct {
		grpc   *GRPC    //谷歌的grpc框架
		rpxMap sync.Map //内部自带的rpc存储器
	}
	GRPC struct {
		addr2Conn map[string]*grpc.ClientConn
		addrMutex sync.Mutex
		server    *grpc.Server
		listen    net.Listener
	}
)

// 添加rpc
func (rx *RPC) Put(rpx *Rpx) {
	rpx.del = rx.del
	rx.rpxMap.Store(rpx.RID(), rpx)
}

// 获取RPC
func (rx *RPC) Get(id uint32) *Rpx {
	if dr, ok := rx.rpxMap.Load(id); ok {
		return dr.(*Rpx)
	}
	return nil
}

// 删除rpc
func (rx *RPC) del(id uint32) {
	if dr, ok := rx.rpxMap.LoadAndDelete(id); ok {
		dr.(*Rpx).release()
	}
}

func (rx *RPC) SetAddr(addr string) {

}

func (rx *RPC) Serve() {
	if rx.grpc != nil && rx.grpc.listen != nil {
		go rx.grpc.server.Serve(rx.grpc.listen)
	}
}

// 开启服务
func (rx *RPC) Start() {
	rpcAddr := gox.Config.RpcAddr
	if rpcAddr != "" {
		rx.grpc.start(rpcAddr)
	}
}

// 停止服务
func (rx *RPC) Stop() {
	if rx.grpc != nil {
		rx.grpc.stop()
	}
}

// 获取grpc 服务端
func (rx *RPC) GRpcServer() *grpc.Server {
	if rx.grpc != nil {
		return rx.grpc.server
	}
	return nil
}

// 获取grpc客户端
func (rx *RPC) GetClientConnByAddr(addr string) *grpc.ClientConn {
	return rx.grpc.getConnByAddr(addr)
}

// 初始化
func New() *RPC {
	return &RPC{
		grpc: &GRPC{
			addr2Conn: make(map[string]*grpc.ClientConn),
		},
	}
}

// start 开启服务
func (grx *GRPC) start(addr string) {
	if grx.listen == nil {
		var err error
		grx.listen, err = net.Listen("tcp", addr)
		if err != nil {
			logger.Fatal().Err(err).Msg("gprc 监听失败")
		}
		grx.server = grpc.NewServer()
		logger.Info().Str("RpcAddr", addr).Msg("gprc 等待客户端连接...")
	}
}

// stop 停止服务
func (grx *GRPC) stop() {
	if grx.listen != nil {
		grx.listen.Close()
	}
}

// getConnByAddr 获取grpc客户端
func (grx *GRPC) getConnByAddr(addr string) *grpc.ClientConn {
	grx.addrMutex.Lock()
	defer grx.addrMutex.Unlock()
	conn, ok := grx.addr2Conn[addr]
	if ok {
		return conn
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal().Err(err).Msg("获取grpc客户端失败")
	}
	grx.addr2Conn[addr] = conn
	return conn
}
