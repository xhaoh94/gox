package rpc

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xhaoh94/gox/engine/logger"
)

// Rpcx 自定义rpcdata
type (
	Rpcx struct {
		err      error
		rid      uint32
		errChan  chan error
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

func NewRpcx(ctx context.Context, rpcID uint32, response interface{}) *Rpcx {
	rpcx := pool.Get().(*Rpcx)
	rpcx.errChan = make(chan error)
	rpcx.ctx = ctx
	rpcx.rid = rpcID
	rpcx.response = response
	rpcx.err = nil
	return rpcx
}
func NewErrorRpcx(err error) *Rpcx {
	rpcx := pool.Get().(*Rpcx)
	rpcx.err = err
	return rpcx
}

// Run 调用
func (rpcx *Rpcx) Run(err error) {
	if rpcx.err == nil {
		rpcx.errChan <- err
	}
}

// Await 异步等待
func (rpcx *Rpcx) Await() error {
	if rpcx.err == nil {
		select {
		case <-rpcx.ctx.Done():
			rpcx.err = errors.New("rpx Context Done")
			break
		case rpcx.err = <-rpcx.errChan:
			break
		case <-time.After(time.Second * 3):
			rpcx.err = errors.New("rpx 超时")
			logger.Error().Msg("rpx 超时")
			break
		}
	}
	rpcx.close()
	return rpcx.err
}

func (rpcx *Rpcx) close() {
	if rpcx.errChan != nil {
		close(rpcx.errChan)
	}

	if rpcx.rid != 0 && rpcx.del != nil {
		rpcx.del(rpcx.rid)
	} else {
		rpcx.release()
	}
}

func (rpcx *Rpcx) release() {
	rpcx.rid = 0
	rpcx.err = nil
	rpcx.errChan = nil
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
	return rpcx.rid
}

func AssignID() uint32 {
	return atomic.AddUint32(&rpcxOps, 1)
}
