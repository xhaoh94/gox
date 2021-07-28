package event

import (
	"errors"
	"reflect"
	"sync"
)

//Event 事件
type Event struct {
	funcMap map[interface{}]reflect.Value
	mu      sync.Mutex
}

//New 创建事件实例
func New() *Event {
	return &Event{
		funcMap: make(map[interface{}]reflect.Value),
	}
}

func (evt *Event) Bind(event interface{}, task interface{}) error {
	evt.mu.Lock()
	defer evt.mu.Unlock()
	if _, ok := evt.funcMap[event]; ok {
		return errors.New("event already defined")
	}
	f := reflect.ValueOf(task)
	if f.Type().Kind() != reflect.Func {
		return errors.New("task is not a function")
	}
	evt.funcMap[event] = f
	return nil
}
func (evt *Event) Call(event interface{}, params ...interface{}) ([]reflect.Value, error) {
	f, in, err := evt.read(event, params...)
	if err != nil {
		return nil, err
	}
	return f.Call(in), nil
}

func (evt *Event) UnBind(event interface{}) error {
	evt.mu.Lock()
	defer evt.mu.Unlock()
	if _, ok := evt.funcMap[event]; !ok {
		return errors.New("event not defined")
	}
	delete(evt.funcMap, event)
	return nil
}

func (evt *Event) UnBinds() {
	evt.mu.Lock()
	defer evt.mu.Unlock()
	evt.funcMap = make(map[interface{}]reflect.Value)
}

func (evt *Event) Has(event interface{}) bool {
	evt.mu.Lock()
	defer evt.mu.Unlock()
	_, ok := evt.funcMap[event]
	return ok
}

func (evt *Event) Events() []interface{} {
	evt.mu.Lock()
	defer evt.mu.Unlock()
	events := make([]interface{}, 0)
	for k := range evt.funcMap {
		events = append(events, k)
	}
	return events
}

func (evt *Event) EventCount() int {
	evt.mu.Lock()
	defer evt.mu.Unlock()
	return len(evt.funcMap)
}

func (evt *Event) read(event interface{}, params ...interface{}) (reflect.Value, []reflect.Value, error) {
	evt.mu.Lock()
	task, ok := evt.funcMap[event]
	evt.mu.Unlock()
	if !ok {
		return reflect.Value{}, nil, errors.New("no task found for event")
	}
	numIn := task.Type().NumIn()
	in := make([]reflect.Value, numIn)
	for i := range params {
		if i >= numIn {
			break
		}
		param := params[i]
		in[i] = reflect.ValueOf(param)
	}
	return task, in, nil
}
