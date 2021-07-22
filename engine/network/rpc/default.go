package rpc

import (
	"sync"
	"time"

	"github.com/xhaoh94/gox/util"
)

//DefalutRPC 自定义rpcdata
type (
	//IDefaultRPC rpc
	IDefaultRPC interface {
		Await() bool
	}
	DefalutRPC struct {
		SessionID string
		RPCID     uint32
		C         chan bool
		Response  interface{}
	}
)

//Run 调用
func (nr *DefalutRPC) Run(success bool) {
	nr.C <- success
	close(nr.C)
}

//Await 异步等待
func (nr *DefalutRPC) Await() bool {
	select {
	case r := <-nr.C:
		return r
	case <-time.After(time.Second * 3):
		close(nr.C)
		DelRPC(nr.RPCID)
		return false
	}
}

var (
	rpcMap sync.Map
)

//AssignRPCID 获取RPCID
func AssignRPCID() uint32 {
	s := util.GetUUID()
	rpcid := util.StrToHash(s)
	return rpcid
}

//PutRPC 添加rpc
func PutRPC(id uint32, nr *DefalutRPC) {
	rpcMap.Store(id, nr)
}

//GetRPC 获取RPC
func GetRPC(id uint32) *DefalutRPC {
	if nr, ok := rpcMap.Load(id); ok {
		defer DelRPC(id)
		return nr.(*DefalutRPC)
	}
	return nil
}

//DelRPC 删除rpc
func DelRPC(id uint32) {
	rpcMap.Delete(id)
}

//DelRPCBySessionID 删除RPC
func DelRPCBySessionID(id string) {
	rpcMap.Range(func(k interface{}, v interface{}) bool {
		dr := v.(*DefalutRPC)
		if dr.SessionID == id {
			dr.Run(false)
			rpcMap.Delete(k)
		}
		return true
	})
}
