package gox

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xhaoh94/gox/engine/logger"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xevent"

	yaml "gopkg.in/yaml.v3"
)

var (
	__init      bool
	__start     bool
	Ctx         context.Context
	ctxCancelFn context.CancelFunc
	Config      AppConf
	Event       types.IEvent

	//网络服务
	NetWork types.INetwork
	// 定位系统
	Location   types.ILocationSystem
	mainModule types.IModule
)

func Init(appConfPath string) {
	if __init {
		log.Printf("gox: 重复初始化")
		return
	}
	__init = true
	Ctx, ctxCancelFn = context.WithCancel(context.Background())
	Event = xevent.New()
	Config = loadConf(appConfPath)
	if Config.AppID == 0 {
		log.Printf("gox: AppID 必须大于0")
		return
	}
	logger.Init(Config.LogConfPath, Config.Development)
}

func loadConf(appConfPath string) AppConf {
	AppCfg := AppConf{}
	bytes, err := os.ReadFile(appConfPath)
	if err != nil {
		log.Fatalf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return AppCfg
	}
	err = yaml.Unmarshal(bytes, &AppCfg)
	if err != nil {
		log.Fatalf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return AppCfg
	}
	return AppCfg
}

// Run 启动
func Run() {
	if __start {
		return
	}
	__start = true
	if mainModule == nil {
		logger.Fatal().Msg("gox: 没有设置主模块")
		return
	}
	logger.Info().Uint("ID", Config.AppID).Str("Type", Config.AppType).Str("Version", Config.Version).Msg("服务启动")
	logger.Info().Msgf("[ByteOrder:%s]", Config.Network.Endian)
	NetWork.Init()
	mainModule.Init(mainModule)
	NetWork.Start()
	mainModule.Start(mainModule)
	logger.Info().Msg("服务启动成功")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	shutdown()
	os.Exit(1)
}

// Shutdown 关闭
func shutdown() {
	__start = false
	ctxCancelFn()
	mainModule.Destroy(mainModule)
	NetWork.Destroy()
	logger.Info().Msgf("服务退出[sid:%d]", Config.AppID)
}

////////////////////////////////////////////////////////////////

// SetModule 设置网络模块
func SetNetWork(network types.INetwork) {
	NetWork = network
}

// SetModule 设置初始模块
func SetModule(module types.IModule) {
	mainModule = module
}

////////////////////////////////////////////////////////////
