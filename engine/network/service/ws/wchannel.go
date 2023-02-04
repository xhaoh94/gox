package ws

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/gorilla/websocket"
)

var channelPool sync.Pool = sync.Pool{New: func() interface{} { return &WChannel{} }}

type (
	//WChannel TCP信道
	WChannel struct {
		service.Channel
		connGuard sync.RWMutex
		conn      *websocket.Conn
	}
)

func (channe *WChannel) init(conn *websocket.Conn) {
	channe.conn = conn
	channe.Init(channe.write, channe.Conn().RemoteAddr().String(), channe.Conn().LocalAddr().String())
}

// Conn 获取通信体
func (channe *WChannel) Conn() *websocket.Conn {
	channe.connGuard.RLock()
	defer channe.connGuard.RUnlock()
	return channe.conn
}

// Start 开启异步接收数据
func (channe *WChannel) Start() {
	channe.Wg.Add(1)
	go func() {
		defer channe.OnStop()
		channe.Wg.Wait()
	}()
	channe.IsRun = true
	go channe.recvAsync()
}
func (channe *WChannel) recvAsync() {
	defer channe.Wg.Done()
	readTimeout := gox.AppConf.Network.ReadTimeout
	if readTimeout > 0 {
		if err := channe.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil { // timeout
			xlog.Info("websocket addr[%s] 接受数据超时err:[%v]", channe.RemoteAddr(), err)
			channe.Stop() //超时断开链接
		}
	}
	for channe.Conn() != nil && channe.IsRun {
		_, r, err := channe.Conn().NextReader()
		if err != nil {
			xlog.Info("websocket addr[%s] 接受数据超时err:[%v]", channe.RemoteAddr(), err)
			channe.Stop() //超时断开链接
			break
		}
		if channe.Read(r) {
			channe.Stop()
		}
		if channe.IsRun && readTimeout > 0 {
			if err = channe.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil { // timeout
				xlog.Info("websocket addr[%s] 接受数据超时err:[%v]", channe.RemoteAddr(), err)
				channe.Stop() //超时断开链接
			}
		}
	}
}

func (channe *WChannel) write(buf []byte) {
	err := channe.Conn().WriteMessage(gox.AppConf.WebSocket.WebSocketMessageType, buf)
	if err != nil {
		xlog.Error("websocket addr[%s]信道写入失败err:[%v]", channe.RemoteAddr(), err)
	}
}

// Stop 停止信道
func (channe *WChannel) Stop() {
	if !channe.IsRun {
		return
	}
	channe.Conn().Close()
	channe.IsRun = false
}

// OnStop 关闭
func (channe *WChannel) OnStop() {
	channe.Channel.OnStop()
	channe.conn = nil
	channelPool.Put(channe)
}
