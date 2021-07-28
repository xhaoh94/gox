package actor

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/xhaoh94/gox/engine/etcd"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

var (
	actorPool *sync.Pool
)

func init() {
	actorPool = &sync.Pool{
		New: func() interface{} {
			return &actorReg{}
		},
	}
}
func New(engine types.IEngine) *Actor {
	return &Actor{
		engine:        engine,
		actorPrefix:   "location/actor",
		keyToActorReg: make(map[string]*actorReg),
	}
}

type Actor struct {
	engine        types.IEngine
	actorPrefix   string
	actorEs       *etcd.EtcdService
	actorRegLock  sync.RWMutex
	keyToActorReg map[string]*actorReg
}

func (atr *Actor) Register(actorID uint32, sessionID string) {
	actor := &actorReg{actorID: actorID, serviceID: atr.engine.GetServiceID(), sessionID: sessionID}
	b, err := json.Marshal(actor)
	if err != nil {
		xlog.Error("RegisterActor Fail err[%v]", err)
		return
	}
	key := fmt.Sprintf(atr.actorPrefix+"/%d", actorID)
	atr.actorEs.Put(key, string(b))
}
func (atr *Actor) Relay(actorID uint32, msgData []byte) {
	key := fmt.Sprintf(atr.actorPrefix+"/%d", actorID)
	atr.actorRegLock.Lock()
	actor, ok := atr.keyToActorReg[key]
	atr.actorRegLock.Unlock()
	if !ok {
		xlog.Error("Actor not found")
		return
	}
	if actor.serviceID != atr.engine.GetServiceID() {
		xlog.Error("Actor serviceID:[%s] Not Equal curServiceID:[%s]", actor.serviceID, atr.engine.GetServiceID())
		return
	}
	session := atr.engine.GetNetWork().GetSessionById(actor.sessionID)
	if session == nil {
		xlog.Error("Actor session not found")
		return
	}
	session.SendData(msgData)
}
func (atr *Actor) Send(actorID uint32, cmd uint32, msg interface{}) {
	key := fmt.Sprintf(atr.actorPrefix+"/%d", actorID)
	atr.actorRegLock.Lock()
	actor, ok := atr.keyToActorReg[key]
	atr.actorRegLock.Unlock()
	if !ok {
		xlog.Error("Actor not found")
		return
	}
	conf := atr.engine.GetNetWork().GetServiceReg().GetServiceConfByID(actor.serviceID)
	if conf == nil {
		xlog.Error("Actor service not found AppID:[%s]", actor.serviceID)
		return
	}
	session := atr.engine.GetNetWork().GetSessionByAddr(conf.GetInteriorAddr())
	if session == nil {
		xlog.Error("Actor session not found")
		return
	}
	session.Actor(actorID, cmd, msg)
}
