package timemgr

import (
	"reflect"
	"sync"
	"time"

	"github.com/xhaoh94/gox/engine/app"
)

var (
	muxSync sync.RWMutex
	fns     []func()

	tick *time.Ticker

	DeltaTime float32
)

func init() {
	fns = make([]func(), 0)
}

func Add(fn func()) {
	muxSync.Lock()
	fns = append(fns, fn)
	muxSync.Unlock()
}
func Remove(fn func()) {
	muxSync.Lock()
	fns = delete(fns, fn)
	muxSync.Unlock()
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
	defer app.Recover()
	for range tick.C {
		if len(fns) == 0 {
			continue
		}
		muxSync.RLock()
		for _, v := range fns {
			go v()
		}
		muxSync.RUnlock()
	}
}
