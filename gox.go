package gox

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xhaoh94/gox/engine/app"
	"github.com/xhaoh94/gox/engine/types"
	"github.com/xhaoh94/gox/engine/xevent"
	"github.com/xhaoh94/gox/engine/xlog"
	yaml "gopkg.in/yaml.v3"
)

var (
	__init      bool
	__start     bool
	Ctx         context.Context
	ctxCancelFn context.CancelFunc
	AppConf     app.AppConf
	Event       types.IEvent

	NetWork  types.INetwork
	Location types.ILocationSystem
	// ActorSystem   types.IActorSystem
	// ServiceSystem types.IServiceSystem

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
	AppConf = loadConf(appConfPath)
	if AppConf.Eid == 0 {
		log.Printf("gox: AppID 必须大于0")
		return
	}
	xlog.Init(AppConf.Log)
}

func loadConf(appConfPath string) app.AppConf {
	AppCfg := app.AppConf{}
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
		xlog.Fatal("gox: 没有设置主模块")
		return
	}
	xlog.Info("服务启动[sid:%d,type:%s,ver:%s]", AppConf.Eid, AppConf.EType, AppConf.Version)
	xlog.Info("[ByteOrder:%s]", AppConf.Network.Endian)
	NetWork.Init()
	mainModule.Init(mainModule)
	NetWork.Start()
	mainModule.Start(mainModule)

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
	xlog.Info("服务退出[sid:%d]", AppConf.Eid)
	xlog.Destroy()
}

////////////////////////////////////////////////////////////////

// SetModule 设置网络模块
func SetNetWork(network types.INetwork) {
	NetWork = network
	Location = network.LocationSystem()
	// ActorSystem = network.ActorSystem()
	// ServiceSystem = network.ServiceSystem()
}

// SetModule 设置初始模块
func SetModule(module types.IModule) {
	mainModule = module
}

////////////////////////////////////////////////////////////
