package kcp

import (
	"net"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
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
			xlog.Fatal("kcp 启动失败 addr:[%s] err:[%v]", service.GetAddr(), err.Error())
			service.Stop()
			return
		}
	}
	xlog.Info("kcp[%s] 等待客户端连接...", service.GetAddr())
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
			xlog.Error("kcp 收受失败[%v]", err.Error())
			break
		}
		xlog.Info("kcp 连接成功[%s]", conn.RemoteAddr().String())
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
		if connCount > gox.AppConf.Network.ReConnectMax {
			xlog.Info("kcp 创建通信信道失败 addr:[%s] err:[%v]", addr, err)
			return nil
		}
		if !service.IsRun || gox.AppConf.Network.ReConnectInterval == 0 {
			return nil
		}
		time.Sleep(gox.AppConf.Network.ReConnectInterval)
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
