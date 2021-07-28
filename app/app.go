package app

import (
	"runtime"
)

// var (
// 	//Version 版本
// 	Version string
// 	//SID 服务id
// 	SID string
// 	//ServiceType 服务类型
// 	ServiceType string
// )

//GetRuntime 运行平台
func GetRuntime() string {
	return runtime.GOOS
}
