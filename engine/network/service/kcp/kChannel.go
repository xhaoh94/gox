package kcp

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/network/service"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xtaci/kcp-go/v5"
)

var channelPool *sync.Pool = &sync.Pool{New: func() interface{} { return &KChannel{} }}

type (
	//KChannel TCP信道
	KChannel struct {
		service.Channel
		connGuard sync.RWMutex
		conn      *kcp.UDPSession
	}
)

func (channel *KChannel) init(conn *kcp.UDPSession) {
	channel.conn = conn
	channel.Init(channel.write, channel.Conn().RemoteAddr().String(), channel.Conn().LocalAddr().String())
}

// Conn 获取通信体
func (channel *KChannel) Conn() *kcp.UDPSession {
	channel.connGuard.RLock()
	defer channel.connGuard.RUnlock()
	return channel.conn
}

// Start 开启异步接收数据
func (channel *KChannel) Start() {
	channel.Wg.Add(1)
	go channel.run()
	channel.IsRun = true
	go channel.recvAsync()
}
func (channel *KChannel) run() {
	defer channel.OnStop()
	channel.Wg.Wait()
}
func (channel *KChannel) recvAsync() {
	defer channel.Wg.Done()
	readTimeout := gox.AppConf.Network.ReadTimeout
	if readTimeout > 0 {
		if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			xlog.Info("kpc addr[%s] 接受数据超时", channel.RemoteAddr())
			xlog.Info("err:[%v]", err)
			channel.Stop()
		}
	}
	for channel.Conn() != nil && channel.IsRun {
		if stop, err := channel.Read(channel.Conn()); stop {
			xlog.Info("kpc addr[%s] err:[%v]", channel.RemoteAddr(), err)
			channel.Stop()
			break
		}

		if channel.IsRun && readTimeout > 0 {
			if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				xlog.Info("kpc addr[%s] 接受数据超时", channel.RemoteAddr())
				xlog.Info("err:[%v]", err)
				channel.Stop()
			}
		}
	}
}

func (channel *KChannel) write(buf []byte) {
	_, err := channel.Conn().Write(buf)
	if err != nil {
		xlog.Error("kcp addr[%s]信道写入失败err:[%v]", channel.RemoteAddr(), err)
	}
}

// Stop 停止信道
func (channel *KChannel) Stop() {
	if !channel.IsRun {
		return
	}
	channel.Conn().Close()
	channel.IsRun = false
}

// OnStop 关闭
func (channel *KChannel) OnStop() {
	channel.Channel.OnStop()
	channel.conn = nil
	channelPool.Put(channel)
}
