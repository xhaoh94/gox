package app

import (
	"encoding/binary"
	"runtime"
)

func GetNetEndian() binary.ByteOrder {
	switch GetAppCfg().Network.NetEndian {
	case "LittleEndian":
		return binary.LittleEndian
	case "BigEndian":
		return binary.BigEndian
	default:
		return binary.LittleEndian
	}
}

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
