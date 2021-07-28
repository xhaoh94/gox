package kcp

import (
	"net"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xtaci/kcp-go/v5"
)

type KService struct {
	service.Service
	listen *kcp.Listener
}

func (ks *KService) Init(addr string, engine types.IEngine) {
	ks.Service.Init(addr, engine)
	ks.Service.ConnectChannelFunc = ks.connectChannel
}

//Start 启动
func (ks *KService) Start() {
	//初始化socket
	if ks.listen == nil {
		var err error
		ks.listen, err = kcp.ListenWithOptions(ks.GetAddr(), nil, 10, 3)
		if err != nil {
			xlog.Error("#kcp.listen failed! addr:[%s] err:[%v]", ks.GetAddr(), err.Error())
			ks.Stop()
			return
		}
	}
	xlog.Info("kcp service Waiting for clients. -> [%s]", ks.GetAddr())
	go ks.accept()
}

func (ks *KService) accept() {
	defer ks.AcceptWg.Done()
	ks.IsRun = true
	ks.AcceptWg.Add(1)
	for {
		conn, err := ks.listen.AcceptKCP()
		if !ks.IsRun {
			break
		}
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			xlog.Error("#kcp.accept failed:[%v]", err.Error())
			break
		}
		xlog.Info("kcp connect success:[%s]", conn.RemoteAddr().String())
		go ks.connection(conn)
	}
}
func (ks *KService) connection(conn *kcp.UDPSession) {
	kChannel := ks.addChannel(conn)
	ks.OnAccept(kChannel)
}
func (ks *KService) addChannel(conn *kcp.UDPSession) *UChannel {
	kChannel := channelPool.Get().(*UChannel)
	kChannel.init(conn)
	return kChannel
}

//connectChannel 链接新信道
func (ks *KService) connectChannel(addr string) types.IChannel {
	var connCount int
	for {
		conn, err := kcp.DialWithOptions(addr, nil, 10, 3)
		if err == nil {
			return ks.addChannel(conn)
		}
		if connCount > app.ReConnectMax {
			xlog.Info("kcp create channel fail addr:[%s] err:[%v]", addr, err)
			return nil
		}
		time.Sleep(app.ReConnectInterval)
		connCount++
		continue
	}
}

//Stop 停止服务
func (ks *KService) Stop() {
	if !ks.IsRun {
		return
	}
	ks.Service.Stop()
	ks.IsRun = false
	ks.CtxCancelFunc()
	ks.listen.Close()
	// 等待线程结束
	ks.AcceptWg.Wait()
}
