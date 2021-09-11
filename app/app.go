package app

import (
	"runtime"
)

//GetRuntime 运行平台
func GetRuntime() string {
	return runtime.GOOS
}

//GetAppCfg 获取app设置
func GetAppCfg() *appConf {
	if appCfg == nil {
		initCfg()
	}
	return appCfg
}

//IsLoadAppCfg 是否加载过配置
func IsLoadAppCfg() bool {
	b := appCfg != nil
	if !b {
		initCfg()
	}
	return b
}
