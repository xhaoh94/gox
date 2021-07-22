package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"
)

var channelPool *sync.Pool

func init() {
	channelPool = &sync.Pool{
		New: func() interface{} {
			return &TChannel{}
		},
	}
}

type (
	//TChannel TCP信道
	TChannel struct {
		service.Channel
		connGuard sync.RWMutex
		conn      *net.Conn
	}
)

func (t *TChannel) init(conn *net.Conn) {
	t.conn = conn
	t.Init(t.write, t.Conn().RemoteAddr().String(), t.Conn().LocalAddr().String())
}

//Conn 获取通信体
func (t *TChannel) Conn() net.Conn {
	t.connGuard.RLock()
	defer t.connGuard.RUnlock()
	return *t.conn
}

//Start 开启异步接收数据
func (t *TChannel) Start() {
	t.Wg.Add(1)
	go func() {
		defer t.OnStop()
		t.Wg.Wait()
	}()
	t.IsRun = true
	go t.recvAsync()
}
func (t *TChannel) recvAsync() {
	defer t.Wg.Done()
	if app.ReadTimeout > 0 {
		if err := t.Conn().SetReadDeadline(time.Now().Add(app.ReadTimeout)); err != nil {
			xlog.Info("setReadDeadline failed:[%v]", err)
			t.Stop()
		}
	}
	for t.Conn() != nil && t.IsRun {
		t.Read(t.Conn(), t.Stop)
		if t.IsRun && app.ReadTimeout > 0 {
			if err := t.Conn().SetReadDeadline(time.Now().Add(app.ReadTimeout)); err != nil {
				xlog.Info("setReadDeadline failed:[%v]", err)
				t.Stop()
			}
		}
	}
}

func (t *TChannel) write(buf []byte) {
	_, err := t.Conn().Write(buf)
	if err != nil {
		xlog.Error(" tcp channel write err: [%v]", err)
	}
}

//Stop 停止信道
func (t *TChannel) Stop() {
	if !t.IsRun {
		return
	}
	t.Conn().Close()
	t.IsRun = false
}

//OnStop 关闭
func (t *TChannel) OnStop() {
	t.Channel.OnStop()
	t.conn = nil
	channelPool.Put(t)
}
