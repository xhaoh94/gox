package app

import (
	"runtime"
)

// 运行平台
func GetRuntime() string {
	return runtime.GOOS
}
