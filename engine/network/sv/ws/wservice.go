package ws

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network/sv"
	"github.com/xhaoh94/gox/types"

	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/gorilla/websocket"
)

//WService WebSocket服务器
type WService struct {
	sv.Service
	upgrader websocket.Upgrader
	sv       *http.Server
	patten   string
	scheme   string
	path     string
}

func (ws *WService) Init(addr string, engine types.IEngine, ctx context.Context) {
	ws.Service.Init(addr, engine, ctx)
	ws.Service.ConnectChannelFunc = ws.connectChannel
}

//Start 启动
func (ws *WService) Start() {
	ws.patten = app.GetAppCfg().WebSocket.WebSocketPattern
	ws.scheme = app.GetAppCfg().WebSocket.WebSocketScheme
	ws.path = app.GetAppCfg().WebSocket.WebSocketPath
	xlog.Debug("patten[%s] scheme[%s] path[%s]", ws.patten, ws.scheme, ws.path)
	mux := http.NewServeMux()
	mux.HandleFunc(ws.patten, ws.wsPage)
	ws.sv = &http.Server{Addr: ws.GetAddr(), Handler: mux}
	ws.upgrader = websocket.Upgrader{
		// ReadBufferSize:  1024,
		// WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	xlog.Info("websocket[%s] 等待客户端连接...", ws.GetAddr())
	go ws.accept()
}
func (ws *WService) accept() {
	defer ws.AcceptWg.Done()
	ws.IsRun = true
	ws.AcceptWg.Add(1)
	if ln, err := net.Listen("tcp", ws.GetAddr()); err != nil {
		xlog.Fatal("websocket 启动失败: [%s]", err.Error())
	} else {
		cf := app.GetAppCfg().WebSocket.CertFile
		kf := app.GetAppCfg().WebSocket.KeyFile
		if cf != "" && kf != "" {
			err = ws.sv.ServeTLS(ln, cf, kf)
		} else {
			err = ws.sv.Serve(ln)
		}
		if err != nil {
			if err == http.ErrServerClosed {
				xlog.Info("websocket 关闭")
			} else {
				xlog.Fatal("websocket 启动失败: [%s]", err.Error())
			}
		}
	}

	// err := ws.sv.ListenAndServe()
	// if err != nil {
	// 	if err == http.ErrServerClosed {
	// 		xlog.Info("websocket 关闭")
	// 	} else {
	// 		xlog.Error("websocket 监听失败: [%s]", err.Error())
	// 	}
	// }
}
func (ws *WService) wsPage(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		xlog.Error("websocket wsPage: [%s]", err.Error())
		return
	}
	xlog.Info("webSocket 连接成功[%s]", conn.RemoteAddr().String())
	go ws.connection(conn)
}

func (ws *WService) connection(conn *websocket.Conn) {
	wChannel := ws.addChannel(conn)
	ws.OnAccept(wChannel)
}
func (ws *WService) addChannel(conn *websocket.Conn) *WChannel {
	wChannel := channelPool.Get().(*WChannel)
	wChannel.init(conn)
	return wChannel
}

//connectChannel 链接新信道
func (ws *WService) connectChannel(addr string) types.IChannel {
	var connCount int
	for {
		u := url.URL{Scheme: ws.scheme, Host: addr, Path: ws.path}
		// var dialer *websocket.Dialer
		dialer := &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		}

		conn, _, err := dialer.Dial(u.String(), nil)
		if err == nil {
			return ws.addChannel(conn)
		}
		if connCount > app.GetAppCfg().Network.ReConnectMax {
			xlog.Info("websocket 创建通信信道失败 addr:[%s] err:[%v]", addr, err)
			return nil
		}
		if !ws.IsRun || app.GetAppCfg().Network.ReConnectInterval == 0 {
			return nil
		}
		time.Sleep(app.GetAppCfg().Network.ReConnectInterval)
		connCount++
		continue
	}

}

//Stop 停止服务
func (ws *WService) Stop() {
	if !ws.IsRun {
		return
	}
	ws.Service.Stop()
	ws.IsRun = false
	ws.sv.Shutdown(ws.Ctx)
	ws.CtxCancelFunc()
	// 等待线程结束
	ws.AcceptWg.Wait()

}
