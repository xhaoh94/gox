package rpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// XRPC 自定义rpcdata
type (
	XRPC struct {
		sid      uint32
		rid      uint32
		c        chan bool
		ctx      context.Context
		response interface{}
		del      func(uint32)
	}
)

var (
	pool   *sync.Pool
	rpcOps uint32
)

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &XRPC{}
		},
	}
}

func NewXRpc(sid uint32, ctx context.Context, response interface{}) *XRPC {
	dr := pool.Get().(*XRPC)
	dr.sid = sid
	dr.c = make(chan bool)
	dr.ctx = ctx
	dr.response = response
	return dr
}

// Run 调用
func (nr *XRPC) Run(success bool) {
	nr.c <- success
}

// Await 异步等待
func (nr *XRPC) Await() bool {
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

func (dr *XRPC) close() {
	close(dr.c)
	if dr.rid != 0 && dr.del != nil {
		dr.del(dr.rid)
	}
}

func (dr *XRPC) release() {
	dr.sid = 0
	dr.rid = 0
	dr.c = nil
	dr.ctx = nil
	dr.response = nil
	dr.del = nil
	pool.Put(dr)
}

func (dr *XRPC) GetResponse() interface{} {
	return dr.response
}

// RID 获取RPCID
func (dr *XRPC) RID() uint32 {
	if dr.rid == 0 {
		dr.rid = AssignID()
	}
	return dr.rid
}

func AssignID() uint32 {
	return atomic.AddUint32(&rpcOps, 1)
}
