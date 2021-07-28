package rpc

import (
	"net"
	"sync"

	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"google.golang.org/grpc"
)

type (
	GRpcServer struct {
		rpcAddr   string
		engine    types.IEngine
		addr2Conn map[string]*grpc.ClientConn
		addrMutex sync.Mutex
		server    *grpc.Server
		listen    net.Listener
	}
)

//获取服务地址
func (g *GRpcServer) GetAddr() string {
	return g.rpcAddr
}

//NewGrpcServer 初始化
func NewGrpcServer(addr string, engine types.IEngine) *GRpcServer {
	return &GRpcServer{
		rpcAddr: addr,
		engine:  engine,
	}
}

//Start 开启服务
func (g *GRpcServer) Start() {
	g.addr2Conn = make(map[string]*grpc.ClientConn)
	if g.listen == nil {
		var err error
		g.listen, err = net.Listen("tcp", g.rpcAddr)
		if err != nil {
			xlog.Fatal("failed to listen: %v", err)
		}
		g.server = grpc.NewServer()
		xlog.Info("rpc service Waiting for clients. -> [%s]", g.rpcAddr)
		g.engine.GetEvent().Bind("_start_engine_ok_", func() {
			if g.listen != nil {
				go g.server.Serve(g.listen)
			}
		})
	}
}

//Stop 停止服务
func (g *GRpcServer) Stop() {
	if g.listen != nil {
		g.listen.Close()
	}
}

//GetServer 获取grpc 服务端
func (g *GRpcServer) GetServer() *grpc.Server {
	return g.server
}

//GetConnByAddr 获取grpc客户端
func (g *GRpcServer) GetConnByAddr(addr string) *grpc.ClientConn {
	defer g.addrMutex.Unlock()
	g.addrMutex.Lock()
	conn, ok := g.addr2Conn[addr]
	if ok {
		return conn
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		xlog.Fatal("did not connect: %v", err)
	}
	g.addr2Conn[addr] = conn
	return conn
}
