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
			return &ActorReg{}
		},
	}
}
func New(engine types.IEngine) *Actor {
	return &Actor{
		engine:        engine,
		actorPrefix:   "location/actor",
		keyToActorReg: make(map[string]*ActorReg),
	}
}

type Actor struct {
	engine        types.IEngine
	actorPrefix   string
	actorEs       *etcd.EtcdService
	actorRegLock  sync.RWMutex
	keyToActorReg map[string]*ActorReg
}

func (atr *Actor) Register(actorID uint32, sessionID string) {
	actor := &ActorReg{ActorID: actorID, ServiceID: atr.engine.GetServiceID(), SessionID: sessionID}
	b, err := json.Marshal(actor)
	if err != nil {
		xlog.Error("注册actor失败err:[%v]", err)
		return
	}
	key := fmt.Sprintf(atr.actorPrefix+"/%d", actorID)
	atr.actorEs.Put(key, string(b))
}
func (atr *Actor) Get(actorID uint32) *ActorReg {
	key := fmt.Sprintf(atr.actorPrefix+"/%d", actorID)
	atr.actorRegLock.Lock()
	actor, ok := atr.keyToActorReg[key]
	atr.actorRegLock.Unlock()
	if !ok {
		xlog.Error("找不到对应的actor。id[%d]", actorID)
		return nil
	}
	return actor
}
func (atr *Actor) Send(actorID uint32, cmd uint32, msg interface{}) bool {
	ar := atr.Get(actorID)
	if ar == nil {
		return false
	}
	conf := atr.engine.GetNetWork().GetServiceReg().GetServiceConfByID(ar.ServiceID)
	if conf == nil {
		xlog.Error("actor找不到服务 ServiceID:[%s]", ar.ServiceID)
		return false
	}
	session := atr.engine.GetNetWork().GetSessionByAddr(conf.GetInteriorAddr())
	if session == nil {
		xlog.Error("actor没有找到session。id[%d]", ar.SessionID)
		return false
	}
	return session.Actor(actorID, cmd, msg)
}
