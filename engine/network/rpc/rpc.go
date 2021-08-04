package rpc

import (
	"net"
	"sync"

	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"

	"google.golang.org/grpc"
)

type (
	RPC struct {
		engine types.IEngine
		grpc   *gRPC    //谷歌的grpc框架
		rpcMap sync.Map //内部自带的rpc存储器
	}
	gRPC struct {
		rpcAddr   string
		addr2Conn map[string]*grpc.ClientConn
		addrMutex sync.Mutex
		server    *grpc.Server
		listen    net.Listener
	}
)

//Put 添加rpc
func (g *RPC) Put(id uint32, dr *DefalutRPC) {
	dr.rid = id
	dr.del = g.del
	g.rpcMap.Store(id, dr)
}

//Get 获取RPC
func (g *RPC) Get(id uint32) *DefalutRPC {
	if dr, ok := g.rpcMap.Load(id); ok {
		return dr.(*DefalutRPC)
	}
	return nil
}

//Del 删除rpc
func (g *RPC) del(id uint32) {
	if dr, ok := g.rpcMap.LoadAndDelete(id); ok {
		dr.(*DefalutRPC).reset()
	}
}

//DelRPCBySessionID 删除RPC
// func (g *RPC) DelRPCBySessionID(id string) {
// 	g.rpcMap.Range(func(k interface{}, v interface{}) bool {
// 		dr := v.(*DefalutRPC)
// 		if dr.sid == id {
// 			dr.Run(false)
// 			g.rpcMap.Delete(k)
// 		}
// 		return true
// 	})
// }

func (g *RPC) GetAddr() string {
	if g.grpc != nil {
		return g.grpc.rpcAddr
	}
	return ""
}

func (g *RPC) Init(addr string) {
	g.grpc = &gRPC{}
	g.engine.GetEvent().Bind("_start_engine_ok_", func() {
		if g.grpc.listen != nil {
			go g.grpc.server.Serve(g.grpc.listen)
		}
	})
}

//Start 开启服务
func (g *RPC) Start() {
	if g.grpc != nil {
		g.grpc.start()
	}
}

//Stop 停止服务
func (g *RPC) Stop() {
	if g.grpc != nil {
		g.grpc.stop()
	}
}

//GetServer 获取grpc 服务端
func (g *RPC) GetServer() *grpc.Server {
	if g.grpc != nil {
		return g.grpc.server
	}
	return nil
}

//GetConnByAddr 获取grpc客户端
func (g *RPC) GetConnByAddr(addr string) *grpc.ClientConn {
	return g.grpc.getConnByAddr(addr)
}

//NewGrpcServer 初始化
func New(engine types.IEngine) *RPC {
	return &RPC{engine: engine}
}

//start 开启服务
func (g *gRPC) start() {
	g.addr2Conn = make(map[string]*grpc.ClientConn)
	if g.listen == nil {
		var err error
		g.listen, err = net.Listen("tcp", g.rpcAddr)
		if err != nil {
			xlog.Fatal("gprc 监听失败: %v", err)
		}
		g.server = grpc.NewServer()
		xlog.Info("grpc[%s] 等待客户端连接...", g.rpcAddr)
	}
}

//stop 停止服务
func (g *gRPC) stop() {
	if g.listen != nil {
		g.listen.Close()
	}
}

//getServer 获取grpc 服务端
func (g *gRPC) getServer() *grpc.Server {
	return g.server
}

//getConnByAddr 获取grpc客户端
func (g *gRPC) getConnByAddr(addr string) *grpc.ClientConn {
	defer g.addrMutex.Unlock()
	g.addrMutex.Lock()
	conn, ok := g.addr2Conn[addr]
	if ok {
		return conn
	}
	var err error
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		xlog.Fatal("获取grpc客户端失败err:[%v]", err)
	}
	g.addr2Conn[addr] = conn
	return conn
}
