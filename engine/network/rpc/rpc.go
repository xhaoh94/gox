package rpc

import (
	"net"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"google.golang.org/grpc"
)

type (
	RPC struct {
		grpc   *GRPC    //谷歌的grpc框架
		rpcMap sync.Map //内部自带的rpc存储器
	}
	GRPC struct {
		addr2Conn map[string]*grpc.ClientConn
		rpcAddr   string
		addrMutex sync.Mutex
		server    *grpc.Server
		listen    net.Listener
	}
)

// Put 添加rpc
func (rpc *RPC) Put(rpcx *Rpcx) {
	rpcx.del = rpc.del
	rpc.rpcMap.Store(rpcx.RID(), rpcx)
}

// Get 获取RPC
func (rpc *RPC) Get(id uint32) *Rpcx {
	if dr, ok := rpc.rpcMap.Load(id); ok {
		return dr.(*Rpcx)
	}
	return nil
}

// Del 删除rpc
func (rpc *RPC) del(id uint32) {
	if dr, ok := rpc.rpcMap.LoadAndDelete(id); ok {
		dr.(*Rpcx).release()
	}
}

func (rpc *RPC) SetAddr(addr string) {

}

func (rpc *RPC) Serve() {
	if rpc.grpc != nil && rpc.grpc.listen != nil {
		go rpc.grpc.server.Serve(rpc.grpc.listen)
	}
}

// Init 开启服务
func (rpc *RPC) Start() {
	rpcAddr := gox.AppConf.RpcAddr
	if rpcAddr != "" {
		rpc.grpc = &GRPC{rpcAddr: rpcAddr}
		rpc.grpc.start()
	}
}

// Destroy 停止服务
func (rpc *RPC) Stop() {
	if rpc.grpc != nil {
		rpc.grpc.stop()
	}
}

// GetServer 获取grpc 服务端
func (rpc *RPC) GetServer() *grpc.Server {
	if rpc.grpc != nil {
		return rpc.grpc.server
	}
	return nil
}

// GetConnByAddr 获取grpc客户端
func (rpc *RPC) GetClientConnByAddr(addr string) *grpc.ClientConn {
	return rpc.grpc.getConnByAddr(addr)
}

// NewGrpcServer 初始化
func New() types.IRPC {
	return &RPC{}
}

// start 开启服务
func (rpc *GRPC) start() {
	rpc.addr2Conn = make(map[string]*grpc.ClientConn)
	if rpc.listen == nil {
		var err error
		rpc.listen, err = net.Listen("tcp", rpc.rpcAddr)
		if err != nil {
			xlog.Fatal("gprc 监听失败: %v", err)
		}
		rpc.server = grpc.NewServer()
		xlog.Info("grpc[%s] 等待客户端连接...", rpc.rpcAddr)
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
		xlog.Fatal("获取grpc客户端失败err:[%v]", err)
	}
	rpc.addr2Conn[addr] = conn
	return conn
}
