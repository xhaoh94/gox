package websocket

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"

	"github.com/gorilla/websocket"
)

var channelPool sync.Pool

func init() {
	channelPool = sync.Pool{
		New: func() interface{} {
			return &WChannel{}
		},
	}
}

type (
	//WChannel TCP信道
	WChannel struct {
		service.Channel
		connGuard sync.RWMutex
		conn      *websocket.Conn
	}
)

func (w *WChannel) init(conn *websocket.Conn) {
	w.conn = conn
	w.Init(w.write, w.Conn().RemoteAddr().String(), w.Conn().LocalAddr().String())
}

//Conn 获取通信体
func (w *WChannel) Conn() *websocket.Conn {
	w.connGuard.RLock()
	defer w.connGuard.RUnlock()
	return w.conn
}

//Start 开启异步接收数据
func (w *WChannel) Start() {
	w.Wg.Add(1)
	go func() {
		defer w.OnStop()
		w.Wg.Wait()
	}()
	w.IsRun = true
	go w.recvAsync()
}
func (w *WChannel) recvAsync() {
	defer w.Wg.Done()
	if app.ReadTimeout > 0 {
		if err := w.Conn().SetReadDeadline(time.Now().Add(app.ReadTimeout)); err != nil { // timeout
			xlog.Info("setReadDeadline failed:[%v] readtimeout[%v]", err, app.ReadTimeout)
			w.Stop() //超时断开链接
		}
	}
	for w.Conn() != nil && w.IsRun {
		_, r, err := w.Conn().NextReader()
		if err != nil {
			xlog.Info("setReadDeadline failed:[%v]", err)
			w.Stop() //超时断开链接
			break
		}
		w.Read(r, w.Stop)
		if w.IsRun && app.ReadTimeout > 0 {
			if err = w.Conn().SetReadDeadline(time.Now().Add(app.ReadTimeout)); err != nil { // timeout
				xlog.Info("setReadDeadline failed:[%v] readtimeout[%v]", err, app.ReadTimeout)
				w.Stop() //超时断开链接
			}
		}
	}
}

func (w *WChannel) write(buf []byte) {
	err := w.Conn().WriteMessage(app.WebSocketMessageType, buf)
	if err != nil {
		xlog.Error("websocket channel write err: [%v]", err)
	}
}

//Stop 停止信道
func (w *WChannel) Stop() {
	if !w.IsRun {
		return
	}
	w.Conn().Close()
	w.IsRun = false
}

//OnStop 关闭
func (w *WChannel) OnStop() {
	w.Channel.OnStop()
	w.conn = nil
	channelPool.Put(w)
}
