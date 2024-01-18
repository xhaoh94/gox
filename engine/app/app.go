package app

import (
	"runtime"

	"github.com/xhaoh94/gox/engine/logger"
)

// 运行平台
func GetRuntime() string {
	return runtime.GOOS
}

func Recover() {
	if err := recover(); err != nil {
		logger.Error().Msgf("recover:%v", err)
	}
}
