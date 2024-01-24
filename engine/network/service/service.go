package service

import (
	"sync"
	"sync/atomic"

	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"
)

type (
	//Service 服务器
	Service struct {
		codec              types.ICodec
		ConnectChannelFunc func(addr string) types.IChannel
		AcceptWg           sync.WaitGroup
		IsRun              bool

		addr           string
		idToSession    map[uint32]*Session //Accept Map
		idMutex        sync.RWMutex
		addrToSession  map[string]*Session //Connect Map
		addrMutex      sync.RWMutex
		sessionWg      sync.WaitGroup
		delSessionFunc func(uint32)
	}
)

var sessionOps uint32
var sessionPool *sync.Pool = &sync.Pool{
	New: func() interface{} {
		return &Session{}
	}}

// Init 服务初始化
func (service *Service) Init(addr string, codec types.ICodec) {
	service.codec = codec
	service.addr = addr
	service.idToSession = make(map[uint32]*Session)
	service.addrToSession = make(map[string]*Session)
}
func (service *Service) Codec() types.ICodec {
	return service.codec
}

// GetAddr 获取地址
func (service *Service) GetAddr() string {
	return service.addr
}

// OnAccept 新链接回调
func (service *Service) OnAccept(channel types.IChannel) {
	session := service.createSession(channel, TagAccept)
	if session != nil {
		service.idMutex.Lock()
		service.idToSession[session.ID()] = session
		service.idMutex.Unlock()
		service.addrMutex.Lock()
		service.addrToSession[session.RemoteAddr()] = session
		service.addrMutex.Unlock()
		session.start()
	}
}

// 通过id获取Session
func (service *Service) GetSessionById(sid uint32) types.ISession {
	defer service.idMutex.RUnlock()
	service.idMutex.RLock()

	if session, ok := service.idToSession[sid]; ok {
		return session
	}
	return nil
}

// 通过addr地址获取Session
func (service *Service) GetSessionByAddr(addr string) types.ISession {

	service.addrMutex.RLock()
	s, ok := service.addrToSession[addr]
	service.addrMutex.RUnlock()

	if ok {
		return s
	}

	session := service.onConnect(addr)
	if session == nil {
		logger.Error().Str("Addr", addr).Msg("创建Session失败")
		return nil
	}
	service.idMutex.Lock()
	service.idToSession[session.ID()] = session
	service.idMutex.Unlock()
	service.addrMutex.Lock()
	service.addrToSession[addr] = session
	service.addrMutex.Unlock()
	session.start()
	return session
}

// Stop 停止服务
func (service *Service) Stop() {
	service.delSessionFunc = nil
	service.idMutex.Lock()
	for k := range service.idToSession {
		service.idToSession[k].stop()
	}
	service.idMutex.Unlock()
	service.sessionWg.Wait()
}

func (service *Service) LinstenByDelSession(callback func(uint32)) {
	service.delSessionFunc = callback
}

func (service *Service) delSession(session types.ISession) {
	if service.delSessionByID(session.ID()) && service.delSessionByAddr(session.RemoteAddr()) {
		if service.delSessionFunc != nil {
			go service.delSessionFunc(session.ID())
		}
		service.sessionWg.Done()
	}
}

func (service *Service) delSessionByID(id uint32) bool {
	defer service.idMutex.Unlock()
	service.idMutex.Lock()
	if _, ok := service.idToSession[id]; ok {
		delete(service.idToSession, id)
		return true
	}
	return false
}

func (service *Service) delSessionByAddr(addr string) bool {
	defer service.addrMutex.Unlock()
	service.addrMutex.Lock()
	if _, ok := service.addrToSession[addr]; ok {
		delete(service.addrToSession, addr)
		return true
	}
	return false
}

func (service *Service) onConnect(addr string) *Session {
	channel := service.ConnectChannelFunc(addr)
	if channel != nil {
		return service.createSession(channel, TagConnector)
	}
	return nil
}

func (service *Service) createSession(channel types.IChannel, tag Tag) *Session {
	sid := atomic.AddUint32(&sessionOps, 1)
	session := sessionPool.Get().(*Session)
	session.init(sid, service, channel, tag)
	if session != nil {
		service.sessionWg.Add(1)
	}
	return session
}
