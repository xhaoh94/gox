package kcp

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xtaci/kcp-go/v5"
)

var channelPool *sync.Pool

func init() {
	channelPool = &sync.Pool{
		New: func() interface{} {
			return &UChannel{}
		},
	}
}

type (
	//UChannel TCP信道
	UChannel struct {
		service.Channel
		connGuard sync.RWMutex
		conn      *kcp.UDPSession
	}
)

func (t *UChannel) init(conn *kcp.UDPSession) {
	t.conn = conn
	t.Init(t.write, t.Conn().RemoteAddr().String(), t.Conn().LocalAddr().String())
}

//Conn 获取通信体
func (t *UChannel) Conn() *kcp.UDPSession {
	t.connGuard.RLock()
	defer t.connGuard.RUnlock()
	return t.conn
}

//Start 开启异步接收数据
func (t *UChannel) Start() {
	t.Wg.Add(1)
	go func() {
		defer t.OnStop()
		t.Wg.Wait()
	}()
	t.IsRun = true
	go t.recvAsync()
}
func (t *UChannel) recvAsync() {
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

func (t *UChannel) write(buf []byte) {
	_, err := t.Conn().Write(buf)
	if err != nil {
		xlog.Error(" tcp channel write err: [%v]", err)
	}
}

//Stop 停止信道
func (t *UChannel) Stop() {
	if !t.IsRun {
		return
	}
	t.Conn().Close()
	t.IsRun = false
}

//OnStop 关闭
func (t *UChannel) OnStop() {
	t.Channel.OnStop()
	t.conn = nil
	channelPool.Put(t)
}
