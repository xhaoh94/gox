package tcp

import (
	"net"
	"time"

	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

// TService TCP服务器
type TService struct {
	service.Service
	listen net.Listener
}

func (service *TService) Init(addr string, codec types.ICodec, engine types.IEngine) {
	service.Service.Init(addr, codec, engine)
	service.Service.ConnectChannelFunc = service.connectChannel
}

// Start 启动
func (service *TService) Start() {
	//初始化socket
	if service.listen == nil {
		var err error
		service.listen, err = net.Listen("tcp", service.GetAddr())
		if err != nil {
			xlog.Fatal("tcp 启动失败 addr:[%s] err:[%v]", service.GetAddr(), err.Error())
			service.Stop()
			return
		}
	}
	xlog.Info("tcp[%s] 等待客户端连接...", service.GetAddr())
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
			xlog.Error("tcp 收受失败[%v]", err.Error())
			break
		}
		xlog.Info("tcp 连接成功[%s]", conn.RemoteAddr().String())
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
		conn, err := net.DialTimeout("tcp", addr, service.Engine.AppConf().Network.ConnectTimeout)
		if err == nil {
			return service.addChannel(&conn)
		}
		if connCount > service.Engine.AppConf().Network.ReConnectMax {
			xlog.Info("tcp 创建通信信道 addr:[%s] err:[%v]", addr, err)
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
func (service *TService) Stop() {
	if !service.IsRun {
		return
	}
	service.Service.Stop()
	service.IsRun = false
	service.CtxCancelFunc()
	service.listen.Close()
	// 等待线程结束
	service.AcceptWg.Wait()
}
