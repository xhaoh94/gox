package consts

import "errors"

var (
	EtcdMutexLockError = errors.New("Etcd Lock Fail")
	CodecError         = errors.New("Message is Nil")
	CallNetError       = errors.New("CallNet Return Param Count Fail")
)
