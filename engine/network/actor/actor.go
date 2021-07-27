package actor

import (
	"encoding/json"
	"fmt"

	"github.com/xhaoh94/gox/app"
	"github.com/xhaoh94/gox/engine/network"
	"github.com/xhaoh94/gox/engine/xlog"
)

var isRun bool

func Start() {
	if isRun {
		return
	}
	isRun = true
	onStart()
}
func Stop() {
	if !isRun {
		return
	}
	onStop()
	isRun = false
}
func Register(actorID uint32, sessionID string) {
	actor := &actorReg{actorID: actorID, serviceID: app.SID, sessionID: sessionID}
	b, err := json.Marshal(actor)
	if err != nil {
		xlog.Error("RegisterActor Fail err[%v]", err)
		return
	}
	key := fmt.Sprintf(actorPrefix+"/%d", actorID)
	actorEs.Put(key, string(b))
}
func Relay(actorID uint32, msgData []byte) {
	key := fmt.Sprintf(actorPrefix+"/%d", actorID)
	actorRegLock.Lock()
	actor, ok := keyToActorReg[key]
	actorRegLock.Unlock()
	if !ok {
		xlog.Error("Actor not found")
		return
	}
	if actor.serviceID != app.SID {
		xlog.Error("Actor serviceID:[%s] Not Equal curServiceID:[%s]", actor.serviceID, app.SID)
		return
	}
	session := network.GetSessionById(actor.sessionID)
	if session == nil {
		xlog.Error("Actor session not found")
		return
	}
	session.SendData(msgData)
}
func Send(actorID uint32, cmd uint32, msg interface{}) {
	key := fmt.Sprintf(actorPrefix+"/%d", actorID)
	actorRegLock.Lock()
	actor, ok := keyToActorReg[key]
	actorRegLock.Unlock()
	if !ok {
		xlog.Error("Actor not found")
		return
	}
	conf := network.GetServiceConfByID(actor.serviceID)
	if conf == nil {
		xlog.Error("Actor service not found AppID:[%s]", actor.serviceID)
		return
	}
	session := network.GetSessionByAddr(conf.InteriorAddr)
	if session == nil {
		xlog.Error("Actor session not found")
		return
	}
	session.Actor(actorID, cmd, msg)
}
