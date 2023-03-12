package service

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/logger"

	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/codechelper"
	"github.com/xhaoh94/gox/engine/network/codec"
	"github.com/xhaoh94/gox/engine/network/location"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	//Session 会话
	Session struct {
		SessionTag
		id            uint32
		service       *Service
		channel       types.IChannel
		ctx           context.Context
		ctxCancelFunc context.CancelFunc
	}
)

const (
	H_B_S        byte = 0x01
	H_B_R        byte = 0x02
	C_S_C        byte = 0x03
	RPC_REQUIRE  byte = 0x04
	RPC_RESPONSE byte = 0x05
)

// 获取id
func (session *Session) ID() uint32 {
	return session.id
}

// 链接地址
func (session *Session) RemoteAddr() string {
	return session.channel.RemoteAddr()
}

// 本地地址
func (session *Session) LocalAddr() string {
	return session.channel.LocalAddr()
}

// 初始化
func (session *Session) init(id uint32, service *Service, channel types.IChannel, t Tag) {
	session.id = id
	session.channel = channel
	session.tag = t
	session.service = service
	session.ctx, session.ctxCancelFunc = context.WithCancel(gox.Ctx)
	session.channel.SetSession(session)
}

// 启动
func (session *Session) start() {
	session.channel.Start()
	if session.IsConnector() { //如果是连接者 启动心跳发送
		go session.onHeartbeat()
	}
}

// 关闭
func (session *Session) stop() {
	if !session.isAct() {
		return
	}
	session.channel.Stop()
}

// 关闭连接
func (session *Session) Close() {
	session.stop()
}

// 发送
func (session *Session) Send(cmd uint32, require any) bool {
	if !session.isAct() {
		return false
	}
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(C_S_C)
	pkt.AppendUint32(cmd)
	if err := pkt.AppendMessage(require, session.codec(cmd)); err != nil {
		return false
	}

	session.sendData(pkt.Data())
	return true
}

// 呼叫
func (session *Session) Call(require any, response any) error {
	cmd := cmdhelper.ToCmd(require, response, 0)
	return session.CallByCmd(cmd, require, response)
}

func (session *Session) CallByCmd(cmd uint32, require any, response any) error {
	if !session.isAct() {
		return errors.New("session not active")
	}
	if cmd == 0 {
		return errors.New("cmd == 0 ")
	}

	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(RPC_REQUIRE)
	pkt.AppendUint32(cmd)
	rpcID := rpc.AssignID()
	pkt.AppendUint32(rpcID)
	if err := pkt.AppendMessage(require, session.codec(cmd)); err != nil {
		return err
	}
	rpx := rpc.NewRpx(session.ctx, rpcID, response)
	session.rpc().Put(rpx)
	session.sendData(pkt.Data())
	err := rpx.Await()
	return err
}

// 回应
func (session *Session) reply(cmd uint32, response any, rpcid uint32) bool {
	if !session.isAct() {
		return false
	}
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(RPC_RESPONSE)
	pkt.AppendUint32(cmd)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(response, session.codec(cmd)); err != nil {
		return false
	}
	session.sendData(pkt.Data())
	return true
}

func (session *Session) sendData(buf []byte) {
	if !session.isAct() {
		return
	}
	session.channel.Send(buf)
}

func (s *Session) isAct() bool {
	return s.id != 0
}

// 心跳
func (session *Session) onHeartbeat() {
	id := session.id
	for session.id != 0 && session.id == id {
		select {
		case <-session.ctx.Done():
			goto end
		case <-time.After(gox.Config.Network.Heartbeat):
			session.sendHeartbeat(H_B_S, 0) //发送空的心跳包
		}
	}
end:
}
func (session *Session) sendHeartbeat(t byte, l uint32) {
	if !session.isAct() {
		return
	}
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(t)
	if l > 0 {
		logger.Debug().Int64("time", time.Now().UnixMilli()).Msg("H_B_S")
		pkt.AppendInt64(time.Now().UnixMilli())
	}
	session.sendData(pkt.Data())
}

func (session *Session) parseReader(r io.Reader) (bool, error) {
	if !session.isAct() {
		return true, errors.New("Session已关闭")
	}
	header := make([]byte, 2)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return true, err
	}
	msglen := codechelper.BytesTo[uint16](header, session.endian())
	if msglen == 0 {
		return true, errors.New("读取到网络空包")
	}

	readMaxLen := gox.Config.Network.ReadMsgMaxLen
	if readMaxLen > 0 && int(msglen) > readMaxLen {
		return true, errors.New("网络包体超出界限")
	}

	buf := make([]byte, msglen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return true, err
	}
	go session.parseMsg(buf)
	return false, nil
}

// parseMsg 解析包
func (session *Session) parseMsg(buf []byte) {
	if !session.isAct() {
		return
	}
	pkt := NewByteArray(buf, session.endian())
	defer pkt.Release()
	// pkt.ReadBytes(8) //8位预留的字节
	// logger.Debug().Bytes("buf", buf).Send()
	switch pkt.ReadOneByte() {
	case H_B_S:
		go session.sendHeartbeat(H_B_R, pkt.RemainLength())
		return
	case C_S_C:
		cmd := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			go session.emitMessage(cmd, nil, 0)
			return
		}
		require := protoreg.GetRequireByCmd(cmd)
		if require == nil {
			logger.Error().Uint32("CMD", cmd).Msg("没有找到注册此协议的结构体")
			return
		}
		if err := pkt.ReadMessage(require, session.codec(cmd)); err != nil {
			logger.Error().Uint32("CMD", cmd).Err(err).Msg("解析网络包体失败")
			return
		}
		go session.emitMessage(cmd, require, 0)
		return
	case RPC_REQUIRE:
		cmd := pkt.ReadUint32()
		rpcID := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		// xlog.Debug("rpcs:cmd:%d,rpcID:%d,msgLen:%d", cmd, rpcID, msgLen)
		if msgLen == 0 {
			go session.emitMessage(cmd, nil, rpcID)
			return
		}
		require := protoreg.GetRequireByCmd(cmd)
		if require == nil {
			logger.Error().Uint32("CMD", cmd).Msg("没有找到注册此协议的结构体")
			go session.reply(cmd, nil, rpcID)
			return
		}
		if err := pkt.ReadMessage(require, session.codec(cmd)); err != nil {
			logger.Error().Uint32("CMD", cmd).Err(err).Msg("解析网络包体失败")
			go session.reply(cmd, nil, rpcID)
			return
		}
		go session.emitMessage(cmd, require, rpcID)
		return
	case RPC_RESPONSE:
		cmd := pkt.ReadUint32()
		rpcID := pkt.ReadUint32()
		rpx := session.rpc().Get(rpcID)
		if rpx != nil {
			msgLen := pkt.RemainLength()
			if msgLen == 0 {
				rpx.Run(errors.New("response len == 0"))
				return
			}
			response := rpx.GetResponse()
			if response == nil {
				rpx.Run(nil)
				return
			}

			if err := pkt.ReadMessage(response, session.codec(cmd)); err != nil {
				logger.Error().Err(err).Msg("解析网络包体失败")
				rpx.Run(err)
				return
			}
			rpx.Run(nil)
		}
		return
	}
}
func (session *Session) Codec() types.ICodec {
	return session.service.Codec()
}

func (session *Session) rpc() *rpc.RPC {
	return gox.NetWork.Rpc().(*rpc.RPC)
}
func (session *Session) codec(cmd uint32) types.ICodec {
	switch cmd {
	case location.LocationGet:
		return codec.MsgPack
	case location.LocationForward:
		return codec.MsgPack
	}
	return session.Codec()
}
func (session *Session) endian() binary.ByteOrder {
	return gox.Config.Network.Endian
}

func (session *Session) emitMessage(cmd uint32, require any, rpcID uint32) {
	if response, err := cmdhelper.CallEvt(cmd, session.ctx, session, require); err == nil {
		if rpcID > 0 {
			session.reply(cmd, response, rpcID)
		}
	} else {
		logger.Warn().Err(err).Msg("Session EmitMessage: 发送消息失败")
	}
}

// release 回收session
func (session *Session) release() {
	logger.Debug().Uint32("ID", session.id).
		Str("Remote", session.RemoteAddr()).Str("Local", session.LocalAddr()).
		Str("Tag", session.GetTagName()).Msg("Session 断开")
	session.ctxCancelFunc()
	session.service.delSession(session)
	session.ctx = nil
	session.ctxCancelFunc = nil
	session.tag = 0
	session.id = 0
	session.channel = nil
	session.service = nil
	sessionPool.Put(session)
}
