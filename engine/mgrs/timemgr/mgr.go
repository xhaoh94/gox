package timemgr

import (
	"reflect"
	"sync"
	"time"
)

var (
	muxSync  sync.RWMutex
	syncFns  []func()
	muxAsync sync.RWMutex
	asyncFns []func()

	tick *time.Ticker

	DeltaTime float32
)

func init() {
	syncFns = make([]func(), 0)
	asyncFns = make([]func(), 0)
}

func Add(fn func(), async bool) {
	if async {
		muxAsync.Lock()
		asyncFns = append(asyncFns, fn)
		muxAsync.Unlock()
	} else {
		muxSync.Lock()
		syncFns = append(syncFns, fn)
		muxSync.Unlock()
	}
}
func Remove(fn func(), async bool) {

	if async {
		muxAsync.Lock()
		asyncFns = delete(asyncFns, fn)
		muxAsync.Unlock()
	} else {
		muxSync.Lock()
		syncFns = delete(syncFns, fn)
		muxSync.Unlock()
	}
}
func delete(list []func(), elem func()) []func() {
	j := 0
	sf1 := reflect.ValueOf(elem)
	for _, v := range list {
		sf2 := reflect.ValueOf(v)
		if sf1.Pointer() != sf2.Pointer() {
			list[j] = v
			j++
		}
	}
	return list[:j]
}

func Start(farme int) {
	DeltaTime = 1 / float32(farme)
	tick = time.NewTicker(time.Second / time.Duration(farme))
	go update()
}
func Stop() {
	tick.Stop()
}
func update() {
	for range tick.C {
		go update_async()
		go update_sync()
	}
}
func update_sync() {
	if len(syncFns) == 0 {
		return
	}
	muxSync.RLock()
	for _, v := range syncFns {
		v()
	}
	muxSync.RUnlock()
}
func update_async() {
	if len(asyncFns) == 0 {
		return
	}
	muxAsync.RLock()
	for _, v := range asyncFns {
		go v()
	}
	muxAsync.RUnlock()
}
