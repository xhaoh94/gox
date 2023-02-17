package network

import (
	"reflect"
	"sync"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	Actor struct {
		fnLock  sync.RWMutex
		fnList  []interface{}
		cmdLock sync.RWMutex
		cmdList []uint32
	}
)

// AddActorFn 添加Actor回调
func (actor *Actor) AddActorFn(fn interface{}) {
	defer actor.fnLock.Unlock()
	actor.fnLock.Lock()
	if actor.fnList == nil {
		actor.fnList = make([]interface{}, 0)
	}
	actor.fnList = append(actor.fnList, fn)
}
func (art *Actor) Init() {
	fnList := actor.GetFnList()
	if fnList == nil {
		xlog.Error("Actor没有注册回调函数")
		return
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := as.parseFn(aid, fn); cmd != 0 {
			actor.SetCmdList(cmd)
		}
	}
}
func (art *Actor) Destroy() {
	art.fnList = nil
	art.cmdList = nil
}

func (art *Actor) GetFnList() []interface{} {
	defer art.fnLock.RUnlock()
	art.fnLock.RLock()
	return art.fnList
}

func (art *Actor) GetCmdList() []uint32 {
	defer art.cmdLock.RUnlock()
	art.cmdLock.RLock()
	return art.cmdList
}
func (art *Actor) SetCmdList(cmd uint32) {
	defer art.cmdLock.Unlock()
	art.cmdLock.Lock()
	if art.cmdList == nil {
		art.cmdList = make([]uint32, 0)
	}
	art.cmdList = append(art.cmdList, cmd)
}

func (as *Actor) parseFn(aid uint32, fn interface{}) uint32 {
	tVlaue := reflect.ValueOf(fn)
	tFun := tVlaue.Type()
	if tFun.Kind() != reflect.Func {
		xlog.Error("Actor回调不是方法")
		return 0
	}
	var out reflect.Type
	switch tFun.NumOut() {
	case 0: //存在没有返回参数的情况
		break
	case 1:
		out = tFun.Out(0)
	default:
		xlog.Error("Actor回调参数有误")
	}
	var in reflect.Type
	switch tFun.NumIn() {
	case 1: //ctx
		break
	case 2: //ctx,req 或 ctx,session
		if out != nil {
			in = tFun.In(1)
		}
	case 3: // ctx,session,req
		if tFun.NumOut() == 1 {
			xlog.Error("Actor回调参数有误")
			return 0
		}
		in = tFun.In(2)
	default:
		xlog.Error("Actor回调参数有误")
	}
	if out == nil && in == nil {
		xlog.Error("Actor回调参数有误")
		return 0
	}
	cmd := cmdhelper.ToCmdByRtype(in, out, aid)
	if cmd == 0 {
		xlog.Error("Actor转换cmd错误")
		return cmd
	}
	if in != nil {
		protoreg.RegisterRType(cmd, in)
	}
	gox.Event.Bind(cmd, fn)
	return cmd
}
