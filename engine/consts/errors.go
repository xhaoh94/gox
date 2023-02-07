package consts

import "errors"

var (
	//etcd 锁失败
	Error_1 = errors.New("etcd lock fail")
	//消息体为空
	Error_2 = errors.New("message is nil")
	//callnet 返回参数错误
	Error_3 = errors.New("callnet return param count fail")
	//Session为空
	Error_4 = errors.New("session is nil")
	//使用已关闭的Session
	Error_5 = errors.New("use of closed network connection")
	//网络空包
	Error_6 = errors.New("empty package")
	//网络读取数据超出最大上限
	Error_7 = errors.New("data out of bounds")
)
