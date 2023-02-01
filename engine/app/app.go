package app

import (
	"runtime"
)

// GetRuntime 运行平台
func GetRuntime() string {
	return runtime.GOOS
}
