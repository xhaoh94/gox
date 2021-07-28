package tcp

import (
	"net"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

//TService TCP服务器
type TService struct {
	service.Service
	listen net.Listener
}

func (ts *TService) Init(addr string, engine types.IEngine) {
	ts.Service.Init(addr, engine)
	ts.Service.ConnectChannelFunc = ts.connectChannel
}

//Start 启动
func (ts *TService) Start() {
	//初始化socket
	if ts.listen == nil {
		var err error
		ts.listen, err = net.Listen("tcp", ts.GetAddr())
		if err != nil {
			xlog.Error("#tcp.listen failed! addr:[%s] err:[%v]", ts.GetAddr(), err.Error())
			ts.Stop()
			return
		}
	}
	xlog.Info("tcp service Waiting for clients. -> [%s]", ts.GetAddr())
	go ts.accept()
}
func (ts *TService) accept() {
	defer ts.AcceptWg.Done()
	ts.IsRun = true
	ts.AcceptWg.Add(1)
	for {
		conn, err := ts.listen.Accept()
		if !ts.IsRun {
			break
		}
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			xlog.Error("#tcp.accept failed:[%v]", err.Error())
			break
		}
		xlog.Info("tcp connect success:[%s]", conn.RemoteAddr().String())
		go ts.connection(&conn)
	}
}
func (ts *TService) connection(conn *net.Conn) {
	tchannel := ts.addChannel(conn)
	ts.OnAccept(tchannel)
}
func (ts *TService) addChannel(conn *net.Conn) *TChannel {
	tChannel := channelPool.Get().(*TChannel)
	tChannel.init(conn)
	return tChannel
}

//connectChannel 链接新信道
func (ts *TService) connectChannel(addr string) types.IChannel {
	var connCount int
	for {
		conn, err := net.DialTimeout("tcp", addr, app.ConnectTimeout)
		if err == nil {
			return ts.addChannel(&conn)
		}
		if connCount > app.ReConnectMax {
			xlog.Info("tcp create channel fail addr:[%s] err:[%v]", addr, err)
			return nil
		}
		time.Sleep(app.ReConnectInterval)
		connCount++
		continue
	}
}

//Stop 停止服务
func (ts *TService) Stop() {
	if !ts.IsRun {
		return
	}
	ts.Service.Stop()
	ts.IsRun = false
	ts.CtxCancelFunc()
	ts.listen.Close()
	// 等待线程结束
	ts.AcceptWg.Wait()
}
