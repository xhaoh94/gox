package sv

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xhaoh94/gox/engine/xlog"
	"github.com/xhaoh94/gox/types"
)

type (
	//Service 服务器
	Service struct {
		Engine             types.IEngine
		Codec              types.ICodec
		ConnectChannelFunc func(addr string) types.IChannel
		AcceptWg           sync.WaitGroup
		IsRun              bool
		Ctx                context.Context
		CtxCancelFunc      context.CancelFunc

		addr          string
		idToSession   map[uint32]*Session //Accept Map
		idMutex       sync.Mutex
		addrToSession map[string]*Session //Connect Map
		addrMutex     sync.Mutex
		sessionWg     sync.WaitGroup
		sessionOps    uint32
	}
)

var sessionPool *sync.Pool = &sync.Pool{New: func() interface{} { return &Session{} }}

// Init 服务初始化
func (ser *Service) Init(addr string, codec types.ICodec, engine types.IEngine, ctx context.Context) {
	ser.Engine = engine
	ser.Codec = codec
	ser.Ctx, ser.CtxCancelFunc = context.WithCancel(ctx)
	ser.addr = addr
	ser.idToSession = make(map[uint32]*Session)
	ser.addrToSession = make(map[string]*Session)
}

// GetAddr 获取地址
func (ser *Service) GetAddr() string {
	return ser.addr
}

// OnAccept 新链接回调
func (ser *Service) OnAccept(channel types.IChannel) {
	session := ser.createSession(channel, TagAccept)
	if session != nil {
		ser.idMutex.Lock()
		ser.idToSession[session.ID()] = session
		ser.idMutex.Unlock()
		ser.addrMutex.Lock()
		ser.addrToSession[session.RemoteAddr()] = session
		ser.addrMutex.Unlock()
		session.start()
	}
}

// GetSessionById 通过id获取Session
func (ser *Service) GetSessionById(sid uint32) types.ISession {
	defer ser.idMutex.Unlock()
	ser.idMutex.Lock()
	session, ok := ser.idToSession[sid]
	if ok {
		return session
	}
	return nil
}

// GetSessionByAddr 通过addr地址获取Session
func (ser *Service) GetSessionByAddr(addr string) types.ISession {
	defer ser.addrMutex.Unlock()
	ser.addrMutex.Lock()
	if s, ok := ser.addrToSession[addr]; ok {
		return s
	}
	session := ser.onConnect(addr)
	if session == nil {
		xlog.Error("创建session失败 addr:[%s]", addr)
		return nil
	}
	ser.idMutex.Lock()
	ser.idToSession[session.ID()] = session
	ser.idMutex.Unlock()
	ser.addrToSession[addr] = session
	session.start()
	return session
}

// Stop 停止服务
func (ser *Service) Stop() {
	ser.idMutex.Lock()
	for k := range ser.idToSession {
		ser.idToSession[k].stop()
	}
	ser.idMutex.Unlock()

	// ser.addrMutex.Lock()
	// for k := range ser.addrToSession {
	// 	ser.addrToSession[k].stop()
	// }
	// ser.addrMutex.Unlock()

	ser.sessionWg.Wait()
}

func (ser *Service) delSession(session types.ISession) {
	if ser.delSessionByID(session.ID()) && ser.delSessionByAddr(session.RemoteAddr()) {
		ser.sessionWg.Done()
	}
}

func (ser *Service) delSessionByID(id uint32) bool {
	defer ser.idMutex.Unlock()
	ser.idMutex.Lock()
	if _, ok := ser.idToSession[id]; ok {
		delete(ser.idToSession, id)
		return true
	}
	return false
}

func (ser *Service) delSessionByAddr(addr string) bool {
	defer ser.addrMutex.Unlock()
	ser.addrMutex.Lock()
	if _, ok := ser.addrToSession[addr]; ok {
		delete(ser.addrToSession, addr)
		return true
	}
	return false
}

func (ser *Service) onConnect(addr string) *Session {
	channel := ser.ConnectChannelFunc(addr)
	if channel != nil {
		return ser.createSession(channel, TagConnector)
	}
	return nil
}

func (ser *Service) createSession(channel types.IChannel, tag Tag) *Session {
	sid := atomic.AddUint32(&ser.sessionOps, 1)
	session := sessionPool.Get().(*Session)
	session.init(sid, ser, channel, tag)
	if session != nil {
		ser.sessionWg.Add(1)
	}
	return session
}
