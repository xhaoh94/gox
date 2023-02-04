package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"
)

var channelPool *sync.Pool = &sync.Pool{New: func() interface{} { return &TChannel{} }}

type (
	//TChannel TCP信道
	TChannel struct {
		service.Channel
		connGuard sync.RWMutex
		conn      *net.Conn
	}
)

func (channel *TChannel) init(conn *net.Conn) {
	channel.conn = conn
	channel.Init(channel.write, channel.Conn().RemoteAddr().String(), channel.Conn().LocalAddr().String())
}

// Conn 获取通信体
func (channel *TChannel) Conn() net.Conn {
	channel.connGuard.RLock()
	defer channel.connGuard.RUnlock()
	return *channel.conn
}

// Start 开启异步接收数据
func (channel *TChannel) Start() {
	channel.Wg.Add(1)
	go func() {
		defer channel.OnStop()
		channel.Wg.Wait()
	}()
	channel.IsRun = true
	go channel.recvAsync()
}
func (channel *TChannel) recvAsync() {
	defer channel.Wg.Done()
	readTimeout := gox.AppConf.Network.ReadTimeout
	if readTimeout > 0 {
		if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			xlog.Info("tcp addr[%s] 接受数据超时err:[%v]", channel.RemoteAddr(), err)
			channel.Stop()
		}
	}
	for channel.Conn() != nil && channel.IsRun {
		if channel.Read(channel.Conn()) {
			channel.Stop()
		}
		if channel.IsRun && readTimeout > 0 {
			if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				xlog.Info("tcp addr[%s] 接受数据超时err:[%v]", channel.RemoteAddr(), err)
				channel.Stop()
			}
		}
	}
}

func (channel *TChannel) write(buf []byte) {
	_, err := channel.Conn().Write(buf)
	if err != nil {
		xlog.Error("tcp addr[%s]信道写入失败err:[%v]", channel.RemoteAddr(), err)
	}
}

// Stop 停止信道
func (channel *TChannel) Stop() {
	if !channel.IsRun {
		return
	}
	channel.Conn().Close()
	channel.IsRun = false
}

// OnStop 关闭
func (channel *TChannel) OnStop() {
	channel.Channel.OnStop()
	channel.conn = nil
	channelPool.Put(channel)
}
