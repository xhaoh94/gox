package rpc

import (
	"context"
	"sync"
	"time"
)

//DefalutRPC 自定义rpcdata
type (
	DefalutRPC struct {
		sid      uint32
		rid      uint32
		c        chan bool
		ctx      context.Context
		response interface{}
		del      func(uint32)
	}
)

var (
	pool *sync.Pool
)

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &DefalutRPC{}
		},
	}
}

func NewDefaultRpc(sid uint32, ctx context.Context, response interface{}) *DefalutRPC {
	dr := pool.Get().(*DefalutRPC)
	dr.sid = sid
	dr.c = make(chan bool)
	dr.ctx = ctx
	dr.response = response
	return dr
}

//Run 调用
func (nr *DefalutRPC) Run(success bool) {
	nr.c <- success
}

//Await 异步等待
func (nr *DefalutRPC) Await() bool {
	select {
	case <-nr.ctx.Done():
		nr.close()
		return false
	case r := <-nr.c:
		nr.close()
		return r
	case <-time.After(time.Second * 3):
		nr.close()
		return false
	}
}

func (dr *DefalutRPC) close() {
	close(dr.c)
	if dr.rid != 0 && dr.del != nil {
		dr.del(dr.rid)
	}
}

func (dr *DefalutRPC) reset() {
	dr.sid = 0
	dr.rid = 0
	dr.c = nil
	dr.ctx = nil
	dr.response = nil
	dr.del = nil
	pool.Put(dr)
}

func (dr *DefalutRPC) GetResponse() interface{} {
	return dr.response
}

// var rpcOps uint32

// //AssignID 获取RPCID
// func AssignID() uint32 {
// 	return atomic.AddUint32(&rpcOps, 1)
// }
