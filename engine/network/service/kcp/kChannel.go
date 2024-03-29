package kcp

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/service"
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
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(256, 256)
	// conn.SetMtu(512);
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
	readTimeout := channel.ReadTimeout()
	if readTimeout > 0 {
		if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			logger.Info().Str("Addr", channel.RemoteAddr()).Err(err).Msg("kcp 接受数据超时")
			channel.Stop()
		}
	}
	for channel.Conn() != nil && channel.IsRun {
		if stop, err := channel.Read(channel.Conn()); stop {
			logger.Info().Str("Addr", channel.RemoteAddr()).Err(err).Send()
			channel.Stop()
			break
		}

		if channel.IsRun && readTimeout > 0 {
			if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				logger.Info().Str("Addr", channel.RemoteAddr()).Err(err).Msg("kcp 接受数据超时")
				channel.Stop()
			}
		}
	}
}

func (channel *KChannel) write(buf []byte) {
	_, err := channel.Conn().Write(buf)
	if err != nil {
		logger.Info().Str("Addr", channel.RemoteAddr()).Err(err).Msg("kcp 信道写入失败")
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
