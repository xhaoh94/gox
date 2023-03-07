package ws

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"

	"github.com/xhaoh94/gox/engine/network/service"

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

func (service *WService) Init(addr string, codec types.ICodec) {
	service.Service.Init(addr, codec)
	service.Service.ConnectChannelFunc = service.connectChannel
}

// Start 启动
func (service *WService) Start() {
	service.patten = gox.Config.WebSocket.WebSocketPattern
	service.scheme = gox.Config.WebSocket.WebSocketScheme
	service.path = gox.Config.WebSocket.WebSocketPath
	logger.Debug().Str("patten", service.patten).
		Str("scheme", service.scheme).
		Str("path", service.path).Msg("websocket")
	mux := http.NewServeMux()
	mux.HandleFunc(service.patten, service.wsPage)
	service.sv = &http.Server{Addr: service.GetAddr(), Handler: mux}
	service.upgrader = websocket.Upgrader{
		// ReadBufferSize:  1024,
		// WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	logger.Info().Str("Addr", service.GetAddr()).Msg("websocket 等待客户端连接...")
	go service.accept()
}
func (service *WService) accept() {
	defer service.AcceptWg.Done()
	service.IsRun = true
	service.AcceptWg.Add(1)
	if ln, err := net.Listen("tcp", service.GetAddr()); err != nil {
		logger.Fatal().Err(err).Msg("websocket 启动失败")
	} else {
		cf := gox.Config.WebSocket.CertFile
		kf := gox.Config.WebSocket.KeyFile
		if cf != "" && kf != "" {
			err = service.sv.ServeTLS(ln, cf, kf)
		} else {
			err = service.sv.Serve(ln)
		}
		if err != nil {
			if err == http.ErrServerClosed {
				logger.Info().Err(err).Msg("websocket 关闭")
			} else {
				logger.Fatal().Err(err).Msg("websocket 启动失败")
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
		logger.Error().Err(err).Msg("websocket wsPage")
		return
	}
	logger.Info().Str("RemoteAddr", conn.RemoteAddr().String()).Msg("websocket 连接成功")
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
		if connCount > gox.Config.Network.ReConnectMax {
			logger.Info().Str("RemoteAddr", conn.RemoteAddr().String()).Err(err).Msg("websocket 创建通信信道失败")
			return nil
		}
		if !service.IsRun || gox.Config.Network.ReConnectInterval == 0 {
			return nil
		}
		time.Sleep(gox.Config.Network.ReConnectInterval)
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
	service.sv.Shutdown(gox.Ctx)
	// 等待线程结束
	service.AcceptWg.Wait()

}
