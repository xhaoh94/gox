package kcp

import (
	"net"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xtaci/kcp-go/v5"
)

type KService struct {
	service.Service
	listen *kcp.Listener
}

func (service *KService) Init(addr string, codec types.ICodec) {
	service.Service.Init(addr, codec)
	service.Service.ConnectChannelFunc = service.connectChannel
}

// Start 启动
func (service *KService) Start() {
	//初始化socket
	if service.listen == nil {
		var err error
		service.listen, err = kcp.ListenWithOptions(service.GetAddr(), nil, 10, 3)
		if err != nil {
			logger.Fatal().Str("Addr", service.GetAddr()).Err(err).Msg("kcp 启动失败")
			service.Stop()
			return
		}
	}
	logger.Info().Str("Addr", service.GetAddr()).Msg("kcp 等待客户端连接...")
	go service.accept()
}

func (service *KService) accept() {
	defer service.AcceptWg.Done()
	service.IsRun = true
	service.AcceptWg.Add(1)
	for {
		conn, err := service.listen.AcceptKCP()
		if !service.IsRun {
			break
		}
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			logger.Fatal().Err(err).Msg("kcp 监听客户端连接失败")
			break
		}
		logger.Info().Str("Addr", conn.RemoteAddr().String()).Msg("kcp 连接成功")
		go service.connection(conn)
	}
}
func (service *KService) connection(conn *kcp.UDPSession) {
	kChannel := service.addChannel(conn)
	service.OnAccept(kChannel)
}
func (service *KService) addChannel(conn *kcp.UDPSession) *KChannel {
	kChannel := channelPool.Get().(*KChannel)
	kChannel.init(conn)
	return kChannel
}

// connectChannel 链接新信道
func (service *KService) connectChannel(addr string) types.IChannel {
	var connCount int
	for {
		conn, err := kcp.DialWithOptions(addr, nil, 10, 3)
		if err == nil {
			return service.addChannel(conn)
		}
		if connCount > gox.Config.Network.ReConnectMax {
			logger.Info().Str("Addr", conn.RemoteAddr().String()).Err(err).Msg("kcp 创建通信信道失败")
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
func (service *KService) Stop() {
	if !service.IsRun {
		return
	}
	service.Service.Stop()
	service.IsRun = false
	service.listen.Close()
	// 等待线程结束
	service.AcceptWg.Wait()
}
