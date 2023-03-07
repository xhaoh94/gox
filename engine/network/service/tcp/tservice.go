package tcp

import (
	"net"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/types"
)

// TService TCP服务器
type TService struct {
	service.Service
	listen net.Listener
}

func (service *TService) Init(addr string, codec types.ICodec) {
	service.Service.Init(addr, codec)
	service.Service.ConnectChannelFunc = service.connectChannel
}

// Start 启动
func (service *TService) Start() {
	//初始化socket
	if service.listen == nil {
		var err error
		service.listen, err = net.Listen("tcp", service.GetAddr())
		if err != nil {
			logger.Fatal().Err(err).Str("Addr", service.GetAddr()).Msg("tcp 启动失败")
			service.Stop()
			return
		}
	}
	logger.Info().Str("Addr", service.GetAddr()).Msg("tcp 等待客户端连接...")
	go service.accept()
}
func (service *TService) accept() {
	defer service.AcceptWg.Done()
	service.IsRun = true
	service.AcceptWg.Add(1)
	for {
		conn, err := service.listen.Accept()
		if !service.IsRun {
			break
		}
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			logger.Fatal().Err(err).Msg("tcp 监听客户端连接失败")
			break
		}
		logger.Info().Str("Addr", conn.RemoteAddr().String()).Msg("tcp 连接成功")
		go service.connection(&conn)
	}
}
func (service *TService) connection(conn *net.Conn) {
	tchannel := service.addChannel(conn)
	service.OnAccept(tchannel)
}
func (service *TService) addChannel(conn *net.Conn) *TChannel {
	tChannel := channelPool.Get().(*TChannel)
	tChannel.init(conn)
	return tChannel
}

// connectChannel 链接新信道
func (service *TService) connectChannel(addr string) types.IChannel {
	var connCount int
	for {
		conn, err := net.DialTimeout("tcp", addr, gox.Config.Network.ConnectTimeout)
		if err == nil {
			return service.addChannel(&conn)
		}
		if connCount > gox.Config.Network.ReConnectMax {
			logger.Error().Str("Addr", addr).Err(err).Msg("tcp 创建通信信道失败")
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
func (service *TService) Stop() {
	if !service.IsRun {
		return
	}
	service.Service.Stop()
	service.IsRun = false
	service.listen.Close()
	// 等待线程结束
	service.AcceptWg.Wait()
}
