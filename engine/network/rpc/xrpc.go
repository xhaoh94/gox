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
		sid      uint32
		rid      uint32
		c        chan bool
		ctx      context.Context
		response interface{}
		del      func(uint32)
	}
	Actorx struct {
		rid uint32
		c   chan []byte
		ctx context.Context
		del func(uint32)
	}
)

var (
	pool    *sync.Pool
	rpcxOps uint32

	actorPool *sync.Pool
)

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &Rpcx{}
		},
	}
	actorPool = &sync.Pool{
		New: func() interface{} {
			return &Actorx{}
		},
	}
}

func NewRpcx(sid uint32, ctx context.Context, response interface{}) *Rpcx {
	rpcx := pool.Get().(*Rpcx)
	rpcx.sid = sid
	rpcx.c = make(chan bool)
	rpcx.ctx = ctx
	rpcx.response = response
	return rpcx
}
func NewActorx(ctx context.Context) *Actorx {
	rpcx := actorPool.Get().(*Actorx)
	rpcx.c = make(chan []byte)
	rpcx.ctx = ctx
	return rpcx
}

// Run 调用
func (nr *Rpcx) Run(success bool) {
	nr.c <- success
}

// Await 异步等待
func (nr *Rpcx) Await() bool {
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

func (rpcx *Rpcx) close() {
	close(rpcx.c)
	if rpcx.rid != 0 && rpcx.del != nil {
		rpcx.del(rpcx.rid)
	}
}

func (rpcx *Rpcx) release() {
	rpcx.sid = 0
	rpcx.rid = 0
	rpcx.c = nil
	rpcx.ctx = nil
	rpcx.response = nil
	rpcx.del = nil
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

// Run 调用
func (nr *Actorx) Run(data []byte) {
	nr.c <- data
}

// Await 异步等待
func (nr *Actorx) Await() []byte {
	select {
	case <-nr.ctx.Done():
		nr.close()
		return nil
	case r := <-nr.c:
		nr.close()
		return r
	case <-time.After(time.Second * 3):
		nr.close()
		return nil
	}
}

func (rpcx *Actorx) close() {
	close(rpcx.c)
	if rpcx.rid != 0 && rpcx.del != nil {
		rpcx.del(rpcx.rid)
	}
}

func (rpcx *Actorx) release() {
	rpcx.rid = 0
	rpcx.c = nil
	rpcx.ctx = nil
	rpcx.del = nil
	actorPool.Put(rpcx)
}

// RID 获取RPCID
func (rpcx *Actorx) RID() uint32 {
	if rpcx.rid == 0 {
		rpcx.rid = AssignID()
	}
	return rpcx.rid
}
func AssignID() uint32 {
	return atomic.AddUint32(&rpcxOps, 1)
}
