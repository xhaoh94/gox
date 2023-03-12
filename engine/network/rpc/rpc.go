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
		rpcAddr   string
		addrMutex sync.Mutex
		server    *grpc.Server
		listen    net.Listener
	}
)

// 添加rpc
func (rpc *RPC) Put(rpx *Rpx) {
	rpx.del = rpc.del
	rpc.rpxMap.Store(rpx.RID(), rpx)
}

// 获取RPC
func (rpc *RPC) Get(id uint32) *Rpx {
	if dr, ok := rpc.rpxMap.Load(id); ok {
		return dr.(*Rpx)
	}
	return nil
}

// 删除rpc
func (rpc *RPC) del(id uint32) {
	if dr, ok := rpc.rpxMap.LoadAndDelete(id); ok {
		dr.(*Rpx).release()
	}
}

func (rpc *RPC) SetAddr(addr string) {

}

func (rpc *RPC) Serve() {
	if rpc.grpc != nil && rpc.grpc.listen != nil {
		go rpc.grpc.server.Serve(rpc.grpc.listen)
	}
}

// 开启服务
func (rpc *RPC) Start() {
	rpcAddr := gox.Config.RpcAddr
	if rpcAddr != "" {
		rpc.grpc = &GRPC{rpcAddr: rpcAddr}
		rpc.grpc.start()
	}
}

// 停止服务
func (rpc *RPC) Stop() {
	if rpc.grpc != nil {
		rpc.grpc.stop()
	}
}

// 获取grpc 服务端
func (rpc *RPC) GRpcServer() *grpc.Server {
	if rpc.grpc != nil {
		return rpc.grpc.server
	}
	return nil
}

// 获取grpc客户端
func (rpc *RPC) GetClientConnByAddr(addr string) *grpc.ClientConn {
	return rpc.grpc.getConnByAddr(addr)
}

// 初始化
func New() *RPC {
	return &RPC{}
}

// start 开启服务
func (rpc *GRPC) start() {
	rpc.addr2Conn = make(map[string]*grpc.ClientConn)
	if rpc.listen == nil {
		var err error
		rpc.listen, err = net.Listen("tcp", rpc.rpcAddr)
		if err != nil {
			logger.Fatal().Err(err).Msg("gprc 监听失败")
		}
		rpc.server = grpc.NewServer()
		logger.Info().Str("RpcAddr", rpc.rpcAddr).Msg("gprc 等待客户端连接...")
	}
}

// stop 停止服务
func (rpc *GRPC) stop() {
	if rpc.listen != nil {
		rpc.listen.Close()
	}
}

// getConnByAddr 获取grpc客户端
func (rpc *GRPC) getConnByAddr(addr string) *grpc.ClientConn {
	defer rpc.addrMutex.Unlock()
	rpc.addrMutex.Lock()
	conn, ok := rpc.addr2Conn[addr]
	if ok {
		return conn
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal().Err(err).Msg("获取grpc客户端失败")
	}
	rpc.addr2Conn[addr] = conn
	return conn
}
