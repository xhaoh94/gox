package sv

import (
	"context"
	"encoding/binary"
	"io"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/consts"

	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/engine/network/cmdtool"
	"github.com/xhaoh94/gox/engine/rpc"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
)

type (
	//Session 会话
	Session struct {
		SessionTag
		id            uint32
		sv            *Service
		channel       types.IChannel
		ctx           context.Context
		ctxCancelFunc context.CancelFunc
	}
)

var KEY []byte = []byte("key_key_")

const (
	C_S_C byte = 0x01
	H_B_S byte = 0x02
	H_B_R byte = 0x03
	RPC_S byte = 0x04
	RPC_R byte = 0x05
)

//UID 获取id
func (s *Session) ID() uint32 {
	return s.id
}

//RemoteAddr 链接地址
func (s *Session) RemoteAddr() string {
	return s.channel.RemoteAddr()
}

//LocalAddr 本地地址
func (s *Session) LocalAddr() string {
	return s.channel.LocalAddr()
}

//Init 初始化
func (s *Session) init(id uint32, service *Service, channel types.IChannel, t Tag) {
	s.id = id
	s.channel = channel
	s.tag = t
	s.sv = service
	s.ctx, s.ctxCancelFunc = context.WithCancel(service.Ctx)
	s.channel.SetSession(s)
}

//start 启动
func (s *Session) start() {
	s.channel.Start()
	if s.IsConnector() { //如果是连接者 启动心跳发送
		go s.onHeartbeat()
	}
}

//Stop 关闭
func (s *Session) stop() {
	if !s.isAct() {
		return
	}
	s.channel.Stop()
}

//Close 关闭连接
func (s *Session) Close() {
	s.stop()
}

//Send 发送
func (s *Session) Send(cmd uint32, msg interface{}) bool {
	if !s.isAct() {
		return false
	}
	pkt := NewByteArray(make([]byte, 0), s.endian())
	defer pkt.Release()
	pkt.AppendBytes(KEY)
	pkt.AppendByte(C_S_C)
	pkt.AppendUint32(cmd)
	if err := pkt.AppendMessage(msg, s.codec()); err != nil {
		return false
	}

	s.sendData(pkt.PktData())
	return true
}

//Call 呼叫
func (s *Session) Call(msg interface{}, response interface{}) types.IDefaultRPC {
	dr := rpc.NewDefaultRpc(s.id, s.ctx, response)
	if !s.isAct() {
		defer dr.Run(false)
		return dr
	}
	cmd := cmdtool.ToCmd(msg, response, 0)
	rpcid := s.defaultRpc().AssignID()
	pkt := NewByteArray(make([]byte, 0), s.endian())
	defer pkt.Release()
	pkt.AppendBytes(KEY)
	pkt.AppendByte(RPC_S)
	pkt.AppendUint32(cmd)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(msg, s.codec()); err != nil {
		defer dr.Run(false)
		return dr
	}
	s.defaultRpc().Put(rpcid, dr)
	s.sendData(pkt.PktData())
	return dr
}

func (s *Session) ActorCall(actorID uint32, msg interface{}, response interface{}) types.IDefaultRPC {

	dr := rpc.NewDefaultRpc(s.id, s.ctx, response)
	if !s.isAct() {
		defer dr.Run(false)
		return dr
	}
	if actorID == 0 {
		xlog.Error("ActorCall传入ActorID不能为空")
		defer dr.Run(false)
		return dr
	}

	cmd := cmdtool.ToCmd(msg, response, actorID)
	rpcid := s.defaultRpc().AssignID()
	pkt := NewByteArray(make([]byte, 0), s.endian())
	defer pkt.Release()
	pkt.AppendBytes(KEY)
	pkt.AppendByte(RPC_S)
	pkt.AppendUint32(cmd)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(msg, s.codec()); err != nil {
		defer dr.Run(false)
		return dr
	}
	s.defaultRpc().Put(rpcid, dr)
	s.sendData(pkt.PktData())
	return dr
}

//reply 回应
func (s *Session) reply(msg interface{}, rpcid uint32) bool {
	if !s.isAct() {
		return false
	}
	pkt := NewByteArray(make([]byte, 0), s.endian())
	defer pkt.Release()
	pkt.AppendBytes(KEY)
	pkt.AppendByte(RPC_R)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(msg, s.codec()); err != nil {
		return false
	}
	s.sendData(pkt.PktData())
	return true
}

func (s *Session) sendData(buf []byte) {
	if !s.isAct() {
		return
	}
	xlog.Debug("data %v", buf)
	s.channel.Send(buf)
}

func (s *Session) isAct() bool {
	return s.id != 0
}

//onHeartbeat 心跳
func (s *Session) onHeartbeat() {
	id := s.id
	for s.id != 0 && s.id == id {
		select {
		case <-s.ctx.Done():
			goto end
		case <-time.After(app.GetAppCfg().Network.Heartbeat):
			s.sendHeartbeat(H_B_S) //发送空的心跳包
		}
	}
end:
}
func (s *Session) sendHeartbeat(t byte) {
	if !s.isAct() {
		return
	}
	pkt := NewByteArray(make([]byte, 0), s.endian())
	defer pkt.Release()
	pkt.AppendBytes(KEY)
	pkt.AppendByte(t)
	s.sendData(pkt.PktData())
}

func (s *Session) parseReader(r io.Reader) bool {
	if !s.isAct() {
		return true
	}
	header := make([]byte, 2)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return true
	}
	msgLen := codec.BytesToUint16(header, s.endian())
	if msgLen == 0 {
		xlog.Error("读取到网络空包 local:[%s] remote:[%s]", s.LocalAddr(), s.RemoteAddr())
		return true
	}

	if int(msgLen) > app.GetAppCfg().Network.ReadMsgMaxLen {
		xlog.Error("网络包体超出界限 local:[%s] remote:[%s]", s.LocalAddr(), s.RemoteAddr())
		return true
	}
	buf := make([]byte, msgLen)
	_, err = io.ReadFull(r, buf)

	if err != nil {
		return true
	}
	go s.parseMsg(buf)
	return false
}

//parseMsg 解析包
func (s *Session) parseMsg(buf []byte) {
	if !s.isAct() {
		return
	}
	pkt := NewByteArray(buf, s.endian())
	defer pkt.Release()
	pkt.ReadBytes(8) //8位预留的字节

	t := pkt.ReadOneByte()
	switch t {
	case H_B_S:
		s.sendHeartbeat(H_B_R)
		break
	case C_S_C:
		cmd := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			s.emitMessage(cmd, nil)
			return
		}
		msg := s.network().GetRegProtoMsg(cmd)
		if msg == nil {
			xlog.Error("没有找到注册此协议的结构体 cmd:[%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg, s.codec()); err != nil {
			xlog.Error("解析网络包体失败 cmd:[%d] err:[%v]", cmd, err)
			return
		}
		s.emitMessage(cmd, msg)
		break
	case RPC_S:
		cmd := pkt.ReadUint32()
		rpcID := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			s.emitRpc(cmd, rpcID, nil)
			return
		}
		msg := s.network().GetRegProtoMsg(cmd)
		if msg == nil {
			xlog.Error("没有找到注册此协议的结构体 cmd:[%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg, s.codec()); err != nil {
			xlog.Error("解析网络包体失败 cmd:[%d] err:[%v]", cmd, err)
			return
		}
		s.emitRpc(cmd, rpcID, msg)
		break
	case RPC_R:
		rpcID := pkt.ReadUint32()
		dr := s.defaultRpc().Get(rpcID)
		if dr != nil {
			msgLen := pkt.RemainLength()
			if msgLen == 0 {
				dr.Run(false)
				return
			}
			if err := pkt.ReadMessage(dr.GetResponse(), s.codec()); err != nil {
				xlog.Error("解析网络包体失败 err:[%v]", err)
				dr.Run(false)
				return
			}
			dr.Run(true)
		}
		break
	}
}

func (s *Session) defaultRpc() *rpc.RPC {
	return s.sv.Engine.GetRPC().(*rpc.RPC)
}
func (s *Session) codec() types.ICodec {
	return s.sv.Engine.GetCodec()
}
func (s *Session) network() types.INetwork {
	return s.sv.Engine.GetNetWork()
}
func (s *Session) event() types.IEvent {
	return s.sv.Engine.GetEvent()
}
func (s *Session) endian() binary.ByteOrder {
	return s.sv.Engine.GetEndian()
}

//callEvt 触发
func (s *Session) callEvt(event uint32, params ...interface{}) (interface{}, error) {
	values, err := s.event().Call(event, params...)
	if err != nil {
		return nil, err
	}
	switch len(values) {
	case 0:
		return nil, nil
	case 1:
		return values[0].Interface(), nil
	case 2:
		return values[0].Interface(), (values[1].Interface()).(error)
	default:
		return nil, consts.CallNetError
	}
}

func (s *Session) emitRpc(cmd uint32, rpc uint32, msg interface{}) {
	if r, err := s.callEvt(cmd, s.ctx, msg); err == nil {
		s.reply(r, rpc)
	} else {
		xlog.Warn("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
	}
}

//emitMessage 派发网络消息
func (s *Session) emitMessage(cmd uint32, msg interface{}) {
	if _, err := s.callEvt(cmd, s.ctx, s, msg); err != nil {
		xlog.Warn("发送消息失败cmd:[%d] err:[%v]", cmd, err)
	}
}

//release 回收session
func (s *Session) release() {
	xlog.Info("session 断开 id:[%d] remote:[%s] local:[%s] tag:[%s]", s.id, s.RemoteAddr(), s.LocalAddr(), s.GetTagName())
	s.ctxCancelFunc()
	s.sv.delSession(s)
	// rpc.DelRPCBySessionID(s.id) 现在通过ctx 关闭
	s.ctx = nil
	s.ctxCancelFunc = nil
	s.tag = 0
	s.id = 0
	s.channel = nil
	s.sv = nil
	sessionPool.Put(s)
}
