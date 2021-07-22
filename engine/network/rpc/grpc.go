package rpc

import (
	"net"
	"sync"

	"github.com/xhaoh94/gox/engine/event"
	"github.com/xhaoh94/gox/engine/xlog"

	"google.golang.org/grpc"
)

var (
	rpcAddr   string
	addr2Conn map[string]*grpc.ClientConn
	addrMutex sync.Mutex
	server    *grpc.Server
	listen    net.Listener
	isRun     bool
)

//获取服务地址
func GetAddr() string {
	return rpcAddr
}

//Init 初始化
func Init(addr string) {
	rpcAddr = addr
}

//Start 开启服务
func Start() {
	if isRun {
		return
	}
	if rpcAddr == "" {
		return
	}
	isRun = true
	addr2Conn = make(map[string]*grpc.ClientConn)

	if listen == nil {
		var err error
		listen, err = net.Listen("tcp", rpcAddr)
		if err != nil {
			xlog.Fatal("failed to listen: %v", err)
		}
		server = grpc.NewServer()
		xlog.Info("rpc service Waiting for clients. -> [%s]", rpcAddr)
		event.Bind("_init_module_ok_", func() {
			if listen != nil {
				go server.Serve(listen)
			}
		})
	}
}

//Stop 停止服务
func Stop() {
	if !isRun {
		return
	}
	if listen != nil {
		listen.Close()
	}
	isRun = false
}

//GetServer 获取grpc 服务端
func GetServer() *grpc.Server {
	return server
}

//GetConnByAddr 获取grpc客户端
func GetConnByAddr(addr string) *grpc.ClientConn {
	defer addrMutex.Unlock()
	addrMutex.Lock()
	conn, ok := addr2Conn[addr]
	if ok {
		return conn
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		xlog.Fatal("did not connect: %v", err)
	}
	addr2Conn[addr] = conn
	return conn
}
