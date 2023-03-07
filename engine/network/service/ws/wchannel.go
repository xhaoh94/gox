package ws

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/network/service"

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

func (channel *WChannel) init(conn *websocket.Conn) {
	channel.conn = conn
	channel.Init(channel.write, channel.Conn().RemoteAddr().String(), channel.Conn().LocalAddr().String())
}

// Conn 获取通信体
func (channel *WChannel) Conn() *websocket.Conn {
	channel.connGuard.RLock()
	defer channel.connGuard.RUnlock()
	return channel.conn
}

// Start 开启异步接收数据
func (channel *WChannel) Start() {
	channel.Wg.Add(1)
	go channel.run()
	channel.IsRun = true
	go channel.recvAsync()
}
func (channel *WChannel) run() {
	defer channel.OnStop()
	channel.Wg.Wait()
}
func (channel *WChannel) recvAsync() {
	defer channel.Wg.Done()
	readTimeout := gox.Config.Network.ReadTimeout
	if readTimeout > 0 {
		if err := channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil { // timeout
			logger.Info().Str("RemoteAddr", channel.RemoteAddr()).Err(err).Msg("websocket 接受数据超时")
			channel.Stop() //超时断开链接
		}
	}
	var stop bool = false
	for channel.Conn() != nil && channel.IsRun {
		_, r, err := channel.Conn().NextReader()
		if err != nil {
			logger.Info().Str("RemoteAddr", channel.RemoteAddr()).Err(err).Send()
			channel.Stop()
			break
		}

		if stop, err = channel.Read(r); stop {
			logger.Info().Str("RemoteAddr", channel.RemoteAddr()).Err(err).Send()
			channel.Stop()
			break
		}
		if channel.IsRun && readTimeout > 0 {
			if err = channel.Conn().SetReadDeadline(time.Now().Add(readTimeout)); err != nil { // timeout
				logger.Info().Str("RemoteAddr", channel.RemoteAddr()).Err(err).Msg("websocket 接受数据超时")
				channel.Stop() //超时断开链接
			}
		}
	}
}

func (channel *WChannel) write(buf []byte) {
	err := channel.Conn().WriteMessage(gox.Config.WebSocket.WebSocketMessageType, buf)
	if err != nil {
		logger.Info().Str("RemoteAddr", channel.RemoteAddr()).Err(err).Msg("websocket 信道写入失败")
	}
}

// Stop 停止信道
func (channel *WChannel) Stop() {
	if !channel.IsRun {
		return
	}
	channel.Conn().Close()
	channel.IsRun = false
}

// OnStop 关闭
func (channel *WChannel) OnStop() {
	channel.Channel.OnStop()
	channel.conn = nil
	channelPool.Put(channel)
}
