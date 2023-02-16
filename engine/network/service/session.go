package service

import (
	"context"
	"encoding/binary"
	"io"
	"time"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/consts"

	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/helper/codechelper"
	"github.com/xhaoh94/gox/engine/network/codec"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
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
	H_B_S byte = 0x01
	H_B_R byte = 0x02
	C_S_C byte = 0x03
	RPC_S byte = 0x04
	RPC_R byte = 0x05
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
func (session *Session) Send(cmd uint32, msg any) bool {
	if !session.isAct() {
		return false
	}
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(C_S_C)
	pkt.AppendUint32(cmd)
	if err := pkt.AppendMessage(msg, session.codec(cmd)); err != nil {
		return false
	}

	session.sendData(pkt.Data())
	return true
}

// 呼叫
func (session *Session) Call(msg any, response any) types.IRpcx {
	cmd := cmdhelper.ToCmd(msg, response, 0)
	return session.CallByCmd(cmd, msg, response)
}

func (session *Session) CallByCmd(cmd uint32, msg any, response any) types.IRpcx {
	if !session.isAct() {
		return rpc.NewEmptyRpcx()
	}
	if cmd == 0 {
		return rpc.NewEmptyRpcx()
	}
	rpcx := rpc.NewRpcx(session.ctx, response)
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(RPC_S)
	pkt.AppendUint32(cmd)
	pkt.AppendUint32(rpcx.RID())
	if err := pkt.AppendMessage(msg, session.codec(cmd)); err != nil {
		defer rpcx.Run(false)
		return rpcx
	}
	session.rpc().Put(rpcx)
	session.sendData(pkt.Data())
	return rpcx
}

// 回应
func (session *Session) reply(cmd uint32, msg any, rpcid uint32) bool {
	if !session.isAct() {
		return false
	}
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	pkt.AppendByte(RPC_R)
	pkt.AppendUint32(cmd)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(msg, session.codec(cmd)); err != nil {
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
		case <-time.After(gox.AppConf.Network.Heartbeat):
			session.sendHeartbeat(H_B_S) //发送空的心跳包
		}
	}
end:
}
func (session *Session) sendHeartbeat(t byte) {
	if !session.isAct() {
		return
	}
	pkt := NewByteArray(make([]byte, 0), session.endian())
	defer pkt.Release()
	// pkt.AppendBytes(KEY)
	pkt.AppendByte(t)
	session.sendData(pkt.Data())
}

func (session *Session) parseReader(r io.Reader) (bool, error) {
	if !session.isAct() {
		return true, consts.Error_5
	}
	header := make([]byte, 2)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return true, err
	}
	msgLen := codechelper.BytesToUint16(header, session.endian())
	if msgLen == 0 {
		xlog.Error("读取到网络空包 local:[%s] remote:[%s]", session.LocalAddr(), session.RemoteAddr())
		return true, consts.Error_6
	}

	if int(msgLen) > gox.AppConf.Network.ReadMsgMaxLen {
		xlog.Error("网络包体超出界限 local:[%s] remote:[%s]", session.LocalAddr(), session.RemoteAddr())
		return true, consts.Error_7
	}
	buf := make([]byte, msgLen)
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

	switch pkt.ReadOneByte() {
	case H_B_S:
		session.sendHeartbeat(H_B_R)
		return
	case C_S_C:
		cmd := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			session.emitMessage(cmd, nil)
			return
		}
		msg := protoreg.GetProtoMsg(cmd)
		if msg == nil {
			xlog.Error("没有找到注册此协议的结构体 cmd:[%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg, session.codec(cmd)); err != nil {
			xlog.Error("解析网络包体失败 cmd:[%d] err:[%v]", cmd, err)
			return
		}
		session.emitMessage(cmd, msg)
		return
	case RPC_S:
		cmd := pkt.ReadUint32()
		rpcID := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			session.emitRpc(cmd, rpcID, nil)
			return
		}
		msg := protoreg.GetProtoMsg(cmd)
		if msg == nil {
			xlog.Error("没有找到注册此协议的结构体 cmd:[%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg, session.codec(cmd)); err != nil {
			xlog.Error("解析网络包体失败 cmd:[%d] err:[%v]", cmd, err)
			return
		}
		session.emitRpc(cmd, rpcID, msg)
		return
	case RPC_R:
		cmd := pkt.ReadUint32()
		rpcID := pkt.ReadUint32()
		dr := session.rpc().Get(rpcID)
		if dr != nil {
			msgLen := pkt.RemainLength()
			if msgLen == 0 {
				dr.Run(false)
				return
			}
			if err := pkt.ReadMessage(dr.GetResponse(), session.codec(cmd)); err != nil {
				xlog.Error("解析网络包体失败 err:[%v]", err)
				dr.Run(false)
				return
			}
			dr.Run(true)
		}
		return
	}
}

func (session *Session) rpc() *rpc.RPC {
	return gox.NetWork.Rpc().(*rpc.RPC)
}
func (session *Session) codec(cmd uint32) types.ICodec {
	switch cmd {
	case consts.LocationLock:
		return codec.MsgPack
	case consts.LocationRefresh:
		return codec.MsgPack
	}
	return session.service.Codec
}
func (session *Session) endian() binary.ByteOrder {
	return gox.AppConf.Network.Endian
}

func (session *Session) emitRpc(cmd uint32, rpc uint32, msg any) {
	if r, err := cmdhelper.CallEvt(cmd, session.ctx, msg); err == nil {
		session.reply(cmd, r, rpc)
	} else {
		xlog.Warn("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
	}
}

// emitMessage 派发网络消息
func (session *Session) emitMessage(cmd uint32, msg any) {
	if _, err := cmdhelper.CallEvt(cmd, session.ctx, session, msg); err != nil {
		xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
	}
}

// release 回收session
func (session *Session) release() {
	xlog.Info("session 断开 id:[%d] remote:[%s] local:[%s] tag:[%s]", session.id, session.RemoteAddr(), session.LocalAddr(), session.GetTagName())
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
