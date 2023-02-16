package xevent

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/xhaoh94/gox/engine/helper/strhelper"
	"github.com/xhaoh94/gox/engine/xlog"
)

// Event 事件
type Event struct {
	bingLock  sync.RWMutex
	bingFnMap map[interface{}]reflect.Value
	onLock    sync.RWMutex
	onFnMap   map[interface{}]map[string]reflect.Value
}

// New 创建事件实例
func New() *Event {
	return &Event{
		bingFnMap: make(map[interface{}]reflect.Value),
		onFnMap:   make(map[interface{}]map[string]reflect.Value),
	}
}
func getKeyName(event interface{}, f reflect.Value) string {
	fnName := runtime.FuncForPC(f.Pointer()).Name()
	eName := strhelper.ValToString(event)
	return eName + "_" + fnName
}

// On 监听事件 回调不可带参数
func (evt *Event) On(event interface{}, task interface{}) {
	evt.onLock.Lock()
	defer evt.onLock.Unlock()
	f := reflect.ValueOf(task)
	if f.Type().Kind() != reflect.Func {
		tip := fmt.Sprintf("监听事件对象不是方法 event:[%v]", event)
		xlog.Error(tip)
		return
	}
	key := getKeyName(event, f)
	fnMap, ok := evt.onFnMap[event]
	if !ok {
		evt.onFnMap[event] = make(map[string]reflect.Value)
		fnMap = evt.onFnMap[event]
	}
	fnMap[key] = f
}
func (evt *Event) Off(event interface{}, task interface{}) {
	evt.onLock.Lock()
	defer evt.onLock.Unlock()

	f := reflect.ValueOf(task)
	if f.Type().Kind() != reflect.Func {
		tip := fmt.Sprintf("取消监听事件对象不是方法 event:[%v]", event)
		xlog.Error(tip)
		return
	}

	fnMap, ok := evt.onFnMap[event]
	if !ok {
		return
	}

	key := getKeyName(event, f)
	delete(fnMap, key)
}

func (evt *Event) Offs(event interface{}) {
	evt.onLock.Lock()
	defer evt.onLock.Unlock()
	_, ok := evt.onFnMap[event]
	if !ok {
		return
	}
	delete(evt.onFnMap, event)
}

// Run 派发事件 不会返回参数
func (evt *Event) Run(event interface{}, params ...interface{}) {
	evt.onLock.RLock()
	fnMap, ok := evt.onFnMap[event]
	evt.onLock.RUnlock()
	if !ok {
		return
	}
	go func() {
		for key := range fnMap {
			fn := fnMap[key]
			numIn := fn.Type().NumIn()
			in := make([]reflect.Value, numIn)
			for i := range params {
				if i >= numIn {
					break
				}
				param := params[i]
				in[i] = reflect.ValueOf(param)
			}
			fn.Call(in)
		}
	}()

}

func (evt *Event) Has(event interface{}, task interface{}) bool {
	evt.onLock.RLock()
	defer evt.onLock.RUnlock()
	f := reflect.ValueOf(task)
	if f.Type().Kind() != reflect.Func {
		tip := fmt.Sprintf("监听事件对象不是方法 event:[%v]", event)
		xlog.Error(tip)
		return false
	}

	fnMap, ok := evt.onFnMap[event]
	if !ok {
		return false
	}
	key := getKeyName(event, f)
	_, ok = fnMap[key]
	return ok
}

// Bind 绑定事件，一个事件只能绑定一个回调，回调可带返回参数
func (evt *Event) Bind(event interface{}, task interface{}) error {
	evt.bingLock.Lock()
	defer evt.bingLock.Unlock()
	if _, ok := evt.bingFnMap[event]; ok {
		tip := fmt.Sprintf("重复监听事件 event:[%v]", event)
		xlog.Error(tip)
		return errors.New(tip)
	}
	f := reflect.ValueOf(task)
	if f.Type().Kind() != reflect.Func {
		tip := fmt.Sprintf("监听事件对象不是方法 event:[%v]", event)
		xlog.Error(tip)
		return errors.New(tip)
	}
	evt.bingFnMap[event] = f
	return nil
}
func (evt *Event) UnBind(event interface{}) error {
	evt.bingLock.Lock()
	defer evt.bingLock.Unlock()
	if _, ok := evt.bingFnMap[event]; !ok {
		tip := fmt.Sprintf("没有找到监听的事件 event:[%v]", event)
		return errors.New(tip)
	}
	delete(evt.bingFnMap, event)
	return nil
}

func (evt *Event) UnBinds() {
	evt.bingLock.Lock()
	defer evt.bingLock.Unlock()
	evt.bingFnMap = make(map[interface{}]reflect.Value)
}

// Call 发送事件，存在返回参数
func (evt *Event) Call(event interface{}, params ...interface{}) ([]reflect.Value, error) {
	evt.bingLock.RLock()
	fn, ok := evt.bingFnMap[event]
	evt.bingLock.RUnlock()
	if !ok {
		return nil, errors.New("没有找到监听的事件")
	}
	numIn := fn.Type().NumIn()
	in := make([]reflect.Value, numIn)
	for i := range params {
		if i >= numIn {
			break
		}
		param := params[i]
		in[i] = reflect.ValueOf(param)
	}
	return fn.Call(in), nil
}

func (evt *Event) HasBind(event interface{}) bool {
	evt.bingLock.RLock()
	defer evt.bingLock.RUnlock()
	_, ok := evt.bingFnMap[event]
	return ok
}

func (evt *Event) BindCount() int {
	evt.bingLock.RLock()
	defer evt.bingLock.RUnlock()
	return len(evt.bingFnMap)
}
