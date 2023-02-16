package rpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Rpcx 自定义rpcdata
type (
	Rpcx struct {
		run      bool
		rid      uint32
		c        chan bool
		ctx      context.Context
		response interface{}
		del      func(uint32)
	}
)

var (
	pool    *sync.Pool
	rpcxOps uint32
)

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &Rpcx{}
		},
	}
}

func NewRpcx(ctx context.Context, response interface{}) *Rpcx {
	rpcx := pool.Get().(*Rpcx)
	rpcx.c = make(chan bool)
	rpcx.ctx = ctx
	rpcx.response = response
	rpcx.run = true
	return rpcx
}
func NewEmptyRpcx() *Rpcx {
	rpcx := pool.Get().(*Rpcx)
	rpcx.run = false
	return rpcx
}

// Run 调用
func (rpcx *Rpcx) Run(success bool) {
	if rpcx.run {
		rpcx.c <- success
	}
}

// Await 异步等待
func (rpcx *Rpcx) Await() bool {
	if rpcx.run {
		select {
		case <-rpcx.ctx.Done():
			rpcx.close()
			return false
		case r := <-rpcx.c:
			rpcx.close()
			return r
		case <-time.After(time.Second * 3):
			rpcx.close()
			return false
		}
	}
	return false
}

func (rpcx *Rpcx) close() {
	close(rpcx.c)
	if rpcx.rid != 0 && rpcx.del != nil {
		rpcx.del(rpcx.rid)
	} else {
		rpcx.release()
	}
}

func (rpcx *Rpcx) release() {
	rpcx.rid = 0
	rpcx.c = nil
	rpcx.ctx = nil
	rpcx.response = nil
	rpcx.del = nil
	rpcx.run = false
	pool.Put(rpcx)
}

func (rpcx *Rpcx) GetResponse() interface{} {
	return rpcx.response
}

// RID 获取RPCID
func (rpcx *Rpcx) RID() uint32 {
	if rpcx.rid == 0 {
		rpcx.rid = AssignID()
	}
	return rpcx.rid
}

func AssignID() uint32 {
	return atomic.AddUint32(&rpcxOps, 1)
}
