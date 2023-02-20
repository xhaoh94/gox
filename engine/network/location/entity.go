package location

import (
	"reflect"

	"github.com/xhaoh94/gox"
	"github.com/xhaoh94/gox/engine/helper/cmdhelper"
	"github.com/xhaoh94/gox/engine/network/protoreg"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xlog"
)

type (
	Entity struct {
		fnList  []interface{}
		cmdList []uint32
	}
)

// 添加Actor回调
func (entity *Entity) Register(fn interface{}) {
	if entity.fnList == nil {
		entity.fnList = make([]interface{}, 0)
	}
	entity.fnList = append(entity.fnList, fn)
}

func (entity *Entity) Init(actor types.ILocationEntity) bool {
	actor.OnInit()
	fnList := entity.fnList
	if fnList == nil {
		xlog.Error("Actor没有注册回调函数")
		return false
	}

	if entity.cmdList == nil {
		entity.cmdList = make([]uint32, 0)
	}
	for index := range fnList {
		fn := fnList[index]
		if cmd := entity.parseFn(actor.LocationID(), fn); cmd != 0 {
			entity.cmdList = append(entity.cmdList, cmd)
		}
	}
	return true
}
func (entity *Entity) Destroy() {

	cmdList := entity.cmdList
	for index := range cmdList {
		cmd := cmdList[index]
		protoreg.UnRegisterRType(cmd)
		gox.Event.UnBind(cmd)
	}
	entity.fnList = nil
	entity.cmdList = nil
}

func (entity *Entity) parseFn(aid uint32, fn interface{}) uint32 {
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
