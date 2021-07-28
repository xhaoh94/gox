package service

import (
	"context"
	"sync"

	"github.com/xhaoh94/gox/consts"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	//Service 服务器
	Service struct {
		addr          string
		idToSession   map[string]*Session //Accept Map
		idMutex       sync.Mutex
		addrToSession map[string]*Session //Connect Map
		addrMutex     sync.Mutex
		sessionWg     sync.WaitGroup
		engine        types.IEngine

		ConnectChannelFunc func(addr string) types.IChannel
		AcceptWg           sync.WaitGroup
		IsRun              bool
		Ctx                context.Context
		CtxCancelFunc      context.CancelFunc
	}
)

var (
	sessionPool *sync.Pool
)

func init() {
	sessionPool = &sync.Pool{
		New: func() interface{} {
			return &Session{}
		},
	}
}

//Init 服务初始化
func (ser *Service) Init(addr string, engine types.IEngine) {
	ser.Ctx, ser.CtxCancelFunc = context.WithCancel(context.TODO())
	ser.addr = addr
	ser.engine = engine
	ser.idToSession = make(map[string]*Session)
	ser.addrToSession = make(map[string]*Session)
}

//GetAddr 获取地址
func (ser *Service) GetAddr() string {
	return ser.addr
}

//OnAccept 新链接回调
func (ser *Service) OnAccept(channel types.IChannel) {
	session := ser.createSession(channel, consts.Accept)
	if session != nil {
		ser.idMutex.Lock()
		ser.idToSession[session.UID()] = session
		ser.idMutex.Unlock()
		ser.addrMutex.Lock()
		ser.addrToSession[session.RemoteAddr()] = session
		ser.addrMutex.Unlock()
		session.start()
	}
}

//GetSessionById 通过id获取Session
func (ser *Service) GetSessionById(sid string) types.ISession {
	defer ser.idMutex.Unlock()
	ser.idMutex.Lock()
	session, ok := ser.idToSession[sid]
	if ok {
		return session
	}
	return nil
}

//GetSessionByAddr 通过addr地址获取Session
func (ser *Service) GetSessionByAddr(addr string) types.ISession {
	defer ser.addrMutex.Unlock()
	ser.addrMutex.Lock()
	if s, ok := ser.addrToSession[addr]; ok {
		return s
	}
	session := ser.onConnect(addr)
	if session == nil {
		xlog.Error("create session fail addr:[%s]", addr)
		return nil
	}
	ser.idMutex.Lock()
	ser.idToSession[session.UID()] = session
	ser.idMutex.Unlock()
	ser.addrToSession[addr] = session
	session.start()
	return session
}

//Stop 停止服务
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
	if ser.delSessionByID(session.UID()) && ser.delSessionByAddr(session.RemoteAddr()) {
		ser.sessionWg.Done()
	}
}

func (ser *Service) delSessionByID(id string) bool {
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
		return ser.createSession(channel, consts.Connector)
	}
	return nil
}

func (ser *Service) createSession(channel types.IChannel, tag consts.SessionTag) *Session {
	session := sessionPool.Get().(*Session)
	session.init(ser, channel, tag)
	if session != nil {
		ser.sessionWg.Add(1)
	}
	return session
}
