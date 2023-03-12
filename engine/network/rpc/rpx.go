package rpc

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xhaoh94/gox/engine/logger"
)

// Rpx 自定义rpcdata
type (
	Rpx struct {
		err      error
		rid      uint32
		errChan  chan error
		ctx      context.Context
		response interface{}
		del      func(uint32)
	}
)

var (
	pool   *sync.Pool
	rpxOps uint32
)

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &Rpx{}
		},
	}
}

func NewRpx(ctx context.Context, rpcID uint32, response interface{}) *Rpx {
	rpx := pool.Get().(*Rpx)
	rpx.errChan = make(chan error)
	rpx.ctx = ctx
	rpx.rid = rpcID
	rpx.response = response
	rpx.err = nil
	return rpx
}

// Run 调用
func (rpx *Rpx) Run(err error) {
	if rpx.err == nil {
		rpx.errChan <- err
	}
}

// Await 异步等待
func (rpx *Rpx) Await() error {
	if rpx.err == nil {
		select {
		case <-rpx.ctx.Done():
			rpx.err = errors.New("rpx Context Done")
			break
		case rpx.err = <-rpx.errChan:
			break
		case <-time.After(time.Second * 3):
			rpx.err = errors.New("rpx 超时")
			logger.Error().Msg("rpx 超时")
			break
		}
	}
	rpx.close()
	return rpx.err
}

func (rpx *Rpx) close() {
	if rpx.errChan != nil {
		close(rpx.errChan)
	}

	if rpx.rid != 0 && rpx.del != nil {
		rpx.del(rpx.rid)
	} else {
		rpx.release()
	}
}

func (rpx *Rpx) release() {
	rpx.rid = 0
	rpx.err = nil
	rpx.errChan = nil
	rpx.ctx = nil
	rpx.response = nil
	rpx.del = nil
	pool.Put(rpx)
}

func (rpx *Rpx) GetResponse() interface{} {
	return rpx.response
}

// RID 获取RPCID
func (rpx *Rpx) RID() uint32 {
	return rpx.rid
}

func AssignID() uint32 {
	return atomic.AddUint32(&rpxOps, 1)
}
