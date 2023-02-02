package ws

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/xhaoh94/gox/engine/types"

	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/gorilla/websocket"
)

// WService WebSocket服务器
type WService struct {
	service.Service
	upgrader websocket.Upgrader
	sv       *http.Server
	patten   string
	scheme   string
	path     string
}

func (service *WService) Init(addr string, codec types.ICodec, engine types.IEngine) {
	service.Service.Init(addr, codec, engine)
	service.Service.ConnectChannelFunc = service.connectChannel
}

// Start 启动
func (service *WService) Start() {
	service.patten = service.Engine.AppConf().WebSocket.WebSocketPattern
	service.scheme = service.Engine.AppConf().WebSocket.WebSocketScheme
	service.path = service.Engine.AppConf().WebSocket.WebSocketPath
	xlog.Debug("patten[%s] scheme[%s] path[%s]", service.patten, service.scheme, service.path)
	mux := http.NewServeMux()
	mux.HandleFunc(service.patten, service.wsPage)
	service.sv = &http.Server{Addr: service.GetAddr(), Handler: mux}
	service.upgrader = websocket.Upgrader{
		// ReadBufferSize:  1024,
		// WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	xlog.Info("websocket[%s] 等待客户端连接...", service.GetAddr())
	go service.accept()
}
func (service *WService) accept() {
	defer service.AcceptWg.Done()
	service.IsRun = true
	service.AcceptWg.Add(1)
	if ln, err := net.Listen("tcp", service.GetAddr()); err != nil {
		xlog.Fatal("websocket 启动失败: [%s]", err.Error())
	} else {
		cf := service.Engine.AppConf().WebSocket.CertFile
		kf := service.Engine.AppConf().WebSocket.KeyFile
		if cf != "" && kf != "" {
			err = service.sv.ServeTLS(ln, cf, kf)
		} else {
			err = service.sv.Serve(ln)
		}
		if err != nil {
			if err == http.ErrServerClosed {
				xlog.Info("websocket 关闭")
			} else {
				xlog.Fatal("websocket 启动失败: [%s]", err.Error())
			}
		}
	}

	// err := service.sv.ListenAndServe()
	// if err != nil {
	// 	if err == http.ErrServerClosed {
	// 		xlog.Info("websocket 关闭")
	// 	} else {
	// 		xlog.Error("websocket 监听失败: [%s]", err.Error())
	// 	}
	// }
}
func (service *WService) wsPage(w http.ResponseWriter, r *http.Request) {
	conn, err := service.upgrader.Upgrade(w, r, nil)
	if err != nil {
		xlog.Error("websocket wsPage: [%s]", err.Error())
		return
	}
	xlog.Info("webSocket 连接成功[%s]", conn.RemoteAddr().String())
	go service.connection(conn)
}

func (service *WService) connection(conn *websocket.Conn) {
	wChannel := service.addChannel(conn)
	service.OnAccept(wChannel)
}
func (service *WService) addChannel(conn *websocket.Conn) *WChannel {
	wChannel := channelPool.Get().(*WChannel)
	wChannel.init(conn)
	return wChannel
}

// connectChannel 链接新信道
func (service *WService) connectChannel(addr string) types.IChannel {
	var connCount int
	for {
		u := url.URL{Scheme: service.scheme, Host: addr, Path: service.path}
		// var dialer *websocket.Dialer
		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		}

		conn, _, err := dialer.Dial(u.String(), nil)
		if err == nil {
			return service.addChannel(conn)
		}
		if connCount > service.Engine.AppConf().Network.ReConnectMax {
			xlog.Info("websocket 创建通信信道失败 addr:[%s] err:[%v]", addr, err)
			return nil
		}
		if !service.IsRun || service.Engine.AppConf().Network.ReConnectInterval == 0 {
			return nil
		}
		time.Sleep(service.Engine.AppConf().Network.ReConnectInterval)
		connCount++
		continue
	}

}

// Stop 停止服务
func (service *WService) Stop() {
	if !service.IsRun {
		return
	}
	service.Service.Stop()
	service.IsRun = false
	service.sv.Shutdown(service.Ctx)
	service.CtxCancelFunc()
	// 等待线程结束
	service.AcceptWg.Wait()

}
