package rpc

import (
	"net"
	"sync"

	"github.com/xhaoh94/gox/engine/event"
	"github.com/xhaoh94/gox/engine/xlog"

	"google.golang.org/grpc"
)

type (
	GRpc struct {
		rpcAddr   string
		addr2Conn map[string]*grpc.ClientConn
		addrMutex sync.Mutex
		server    *grpc.Server
		listen    net.Listener
	}
)

var (
	gprc  *GRpc
	isRun bool
)

//获取服务地址
func GetAddr() string {
	return gprc.rpcAddr
}

//Init 初始化
func Init(addr string) {
	gprc = &GRpc{
		rpcAddr: addr,
	}
}

//Start 开启服务
func Start() {
	if isRun {
		return
	}
	if gprc == nil {
		return
	}
	isRun = true
	gprc.addr2Conn = make(map[string]*grpc.ClientConn)

	if gprc.listen == nil {
		var err error
		gprc.listen, err = net.Listen("tcp", gprc.rpcAddr)
		if err != nil {
			xlog.Fatal("failed to listen: %v", err)
		}
		gprc.server = grpc.NewServer()
		xlog.Info("rpc service Waiting for clients. -> [%s]", gprc.rpcAddr)
		event.Bind("_init_module_ok_", func() {
			if gprc.listen != nil {
				go gprc.server.Serve(gprc.listen)
			}
		})
	}
}

//Stop 停止服务
func Stop() {
	if !isRun {
		return
	}
	if gprc == nil {
		return
	}
	if gprc.listen != nil {
		gprc.listen.Close()
	}
	isRun = false
}

//GetServer 获取grpc 服务端
func GetServer() *grpc.Server {
	return gprc.server
}

//GetConnByAddr 获取grpc客户端
func GetConnByAddr(addr string) *grpc.ClientConn {
	defer gprc.addrMutex.Unlock()
	gprc.addrMutex.Lock()
	conn, ok := gprc.addr2Conn[addr]
	if ok {
		return conn
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		xlog.Fatal("did not connect: %v", err)
	}
	gprc.addr2Conn[addr] = conn
	return conn
}
