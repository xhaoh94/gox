package service

import (
	"context"
	"time"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/consts"

	"github.com/xhaoh94/gox/engine/network/actor"
	"github.com/xhaoh94/gox/engine/network/rpc"
	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
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
func (s *Session) ID() string {
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

//Send 发送
func (s *Session) Send(cmd uint32, msg interface{}) bool {
	if !s.isAct() {
		return false
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_csc)
	pkt.AppendUint32(cmd)
	if err := pkt.AppendMessage(msg, s.codec()); err != nil {
		return false
	}
	s.sendData(pkt.PktData())
	return true
}
func (s *Session) Actor(actorID uint32, cmd uint32, msg interface{}) bool {
	if !s.isAct() {
		return false
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_actor)
	pkt.AppendUint32(actorID)
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
	cmd := ToCmd(msg, response)
	rpcid := rpc.AssignID()
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_rpcs)
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

//Reply 回应
func (s *Session) Reply(msg interface{}, rpcid uint32) bool {
	if !s.isAct() {
		return false
	}
	pkt := newByteArray(make([]byte, 0))
	defer pkt.Reset()
	pkt.AppendByte(_rpcr)
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
	s.channel.Send(buf)
}

func (s *Session) isAct() bool {
	return s.id != ""
}

//onHeartbeat 心跳
func (s *Session) onHeartbeat() {
	id := s.id
	for s.id != "" && s.id == id {
		select {
		case <-s.ctx.Done():
			goto end
		case <-time.After(app.GetAppCfg().Network.Heartbeat):
			s.sendHeartbeat(_hbs) //发送空的心跳包
		}
	}
end:
}
func (s *Session) sendHeartbeat(t byte) {
	if !s.isAct() {
		return
	}
	pkg := newByteArray(make([]byte, 0))
	defer pkg.Reset()
	pkg.AppendByte(t)
	s.sendData(pkg.PktData())
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
			xlog.Error("actor 数据长度为0")
			return
		}
		ar := s.network().GetActor().(*actor.Actor).Get(actorID)
		if ar == nil {
			return
		}
		if ar.ServiceID != s.service.engine.GetServiceID() {
			xlog.Error("actor服务id[%s]与当前服务id[%s]不相同", ar.ServiceID, s.service.engine.GetServiceID())
			return
		}
		session := s.service.engine.GetNetWork().GetSessionById(ar.SessionID)
		if session == nil {
			xlog.Error("actor没有找到session。id[%d]", ar.SessionID)
			return
		}
		msgData := pkt.ReadBytes(msgLen)
		temPkt := newByteArray(make([]byte, 0))
		defer temPkt.Reset()
		temPkt.AppendByte(_csc)
		temPkt.AppendBytes(msgData)
		bytes := temPkt.PktData()
		session.(*Session).sendData(bytes)
		break
	case _csc:
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
		msg := s.network().GetRegProtoMsg(cmd)
		if msg == nil {
			xlog.Error("没有找到注册此协议的结构体 cmd:[%d]", cmd)
			return
		}
		if err := pkt.ReadMessage(msg, s.codec()); err != nil {
			xlog.Error("解析网络包体失败 cmd:[%d] err:[%v]", cmd, err)
			return
		}
		go s.emitRpc(cmd, rpcID, msg)
		break
	case _rpcr:
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
	return s.network().GetRPC().(*rpc.RPC)
}
func (s *Session) codec() types.ICodec {
	return s.service.engine.GetCodec()
}
func (s *Session) network() types.INetwork {
	return s.service.engine.GetNetWork()
}
func (s *Session) event() types.IEvent {
	return s.service.engine.GetEvent()
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
		s.Reply(r, rpc)
	} else {
		xlog.Error("发送rpc消息失败cmd:[%d] err:[%v]", cmd, err)
	}
}

//emitMessage 派发网络消息
func (s *Session) emitMessage(cmd uint32, msg interface{}) {
	if _, err := s.callEvt(cmd, s.ctx, s, msg); err != nil {
		xlog.Error("发送消息失败cmd:[%d] err:[%v]", cmd, err)
	}
}

//OnClose 关闭
func (s *Session) close() {
	xlog.Info("session 断开 id:[%s] remote:[%s] local:[%s] tag:[%s]", s.id, s.RemoteAddr(), s.LocalAddr(), s.GetTagName())
	s.ctxCancelFunc()
	s.service.delSession(s)
	s.reset()
	sessionPool.Put(s)
}

func (s *Session) reset() {
	// rpc.DelRPCBySessionID(s.id) 现在通过ctx 关闭
	s.tag = 0
	s.id = ""
	s.channel = nil
	s.service = nil
}
