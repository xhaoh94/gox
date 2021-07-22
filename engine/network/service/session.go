package service

import (
	"context"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/consts"
	"github.com/xhaoh94/gox/engine/event"
	"github.com/xhaoh94/gox/engine/network/actor"
	"github.com/xhaoh94/gox/engine/network/proto"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/network/types"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/util"
)

type (
	//Session 会话
	Session struct {
		SessionTag
		service       *Service
		channel       types.IChannel
		id            string
		ctx           context.Context
		ctxCancelFunc context.CancelFunc
	}
)

const (
	_csc   byte = 0x01
	_hbs   byte = 0x02
	_hbr   byte = 0x03
	_rpcs  byte = 0x04
	_rpcr  byte = 0x05
	_actor byte = 0x06
)

//UID 获取id
func (s *Session) UID() string {
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
func (s *Session) init(service *Service, channel types.IChannel, t consts.SessionTag) {
	s.id = util.GetUUID()
	s.channel = channel
	s.tag = t
	s.service = service
	s.ctx, s.ctxCancelFunc = context.WithCancel(service.Ctx)
	s.channel.SetCallBackFn(s.read, s.close)
}

//start 启动
func (s *Session) start() {
	s.channel.Start()
	if s.GetTag() == consts.Connector { //如果是连接者 启动心跳发送
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

//Send 发送
func (s *Session) Send(cmd uint32, msg interface{}) {
	if !s.isAct() {
		return
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_csc)
	pkt.AppendUint32(cmd)
	if err := pkt.AppendMessage(msg); err != nil {
		return
	}
	s.SendData(pkt.SendData())
}
func (s *Session) Actor(actorID uint32, cmd uint32, msg interface{}) {
	if !s.isAct() {
		return
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_actor)
	pkt.AppendUint32(actorID)
	pkt.AppendUint32(cmd)
	if err := pkt.AppendMessage(msg); err != nil {
		return
	}
	s.SendData(pkt.SendData())
}

//Call 呼叫
func (s *Session) Call(msg interface{}, response interface{}) rpc.IDefaultRPC {
	nr := &rpc.DefalutRPC{SessionID: s.id, C: make(chan bool), Response: response}
	if !s.isAct() {
		defer nr.Run(false)
		return nr
	}
	msgID := converMsgID(msg, response)
	rpcid := rpc.AssignRPCID()
	nr.RPCID = rpcid
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_rpcs)
	pkt.AppendUint32(msgID)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(msg); err != nil {
		defer nr.Run(false)
		return nr
	}
	rpc.PutRPC(rpcid, nr)
	s.SendData(pkt.SendData())
	return nr
}

//Reply 回应
func (s *Session) Reply(msg interface{}, rpcid uint32) {
	if !s.isAct() {
		return
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_rpcr)
	pkt.AppendUint32(rpcid)
	if err := pkt.AppendMessage(msg); err != nil {
		return
	}
	s.SendData(pkt.SendData())
	return
}
func (s *Session) SendData(buf []byte) {
	if !s.isAct() {
		return
	}
	s.channel.Send(buf)
}

func (s *Session) isAct() bool {
	return s.id != ""
}

//onHeartbeat 心跳
func (s *Session) onHeartbeat() {
	id := s.id
	for s.id != "" && s.id == id {
		s.sendHeartbeat(_hbs) //发送空的心跳包
		time.Sleep(app.Heartbeat)
	}
}
func (s *Session) sendHeartbeat(t byte) {
	if !s.isAct() {
		return
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(t)
	s.SendData(pkt.SendData())
}

//OnRead 读取数据
func (s *Session) read(data []byte) {
	if !s.isAct() {
		return
	}
	pkt := newByteArray(data)
	defer pkt.Reset()
	t := pkt.ReadOneByte()
	switch t {
	case _hbs:
		s.sendHeartbeat(_hbr)
		break
	case _actor:
		actorID := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			xlog.Error("actor msglen==0")
			return
		}
		msgData := pkt.ReadBytes(msgLen)
		bytes := []byte{_csc}
		bytes = append(bytes, msgData...)
		go actor.Relay(actorID, bytes)
		break
	case _csc:
		cmd := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			s.emitMessage(cmd, nil)
			return
		}
		msg := proto.GetMsg(cmd)
		if msg == nil {
			xlog.Error("The registered ID was not found [%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg); err != nil {
			xlog.Error("run net event byte2msg mid:[%d] err:[%v]", cmd, err)
			return
		}
		go s.emitMessage(cmd, msg)
		break
	case _rpcs:
		cmd := pkt.ReadUint32()
		rpcID := pkt.ReadUint32()
		msgLen := pkt.RemainLength()
		if msgLen == 0 {
			go s.emitRpc(cmd, rpcID, nil)
			return
		}
		msg := proto.GetMsg(cmd)
		if msg == nil {
			xlog.Error("The registered ID was not found [%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg); err != nil {
			xlog.Error("run net event byte2msg mid:[%d] err:[%v]", cmd, err)
			return
		}
		go s.emitRpc(cmd, rpcID, msg)
		break
	case _rpcr:
		rpcID := pkt.ReadUint32()
		dr := rpc.GetRPC(rpcID)
		if dr != nil {
			msgLen := pkt.RemainLength()
			if msgLen == 0 {
				dr.Run(false)
				return
			}
			if err := pkt.ReadMessage(dr.Response); err != nil {
				xlog.Error("run net event byte2msg err:[%v]", err)
				dr.Run(false)
				return
			}
			dr.Run(true)
		}
		break
	}

}

func (s *Session) emitRpc(cmd uint32, rpc uint32, msg interface{}) {
	if r, err := event.CallNet(cmd, s.ctx, msg); err == nil {
		s.Reply(r, rpc)
	} else {
		xlog.Error("emitRpc err:[%v] cmd:[%d]", err, cmd)
	}
}

//emitMessage 派发网络消息
func (s *Session) emitMessage(cmd uint32, msg interface{}) {
	if _, err := event.CallNet(cmd, s.ctx, s, msg); err != nil {
		xlog.Error("emitMessage err:[%v] cmd:[%d]", err, cmd)
	}
}

//OnClose 关闭
func (s *Session) close() {
	xlog.Info("session close id:[%s] remote:[%s] local:[%s] tag:[%s]", s.id, s.RemoteAddr(), s.LocalAddr(), s.GetTagName())
	s.ctxCancelFunc()
	s.service.delSession(s)
	s.reset()
	sessionPool.Put(s)
}

func (s *Session) reset() {
	rpc.DelRPCBySessionID(s.id)
	s.tag = 0
	s.id = ""
	s.channel = nil
	s.service = nil
}
