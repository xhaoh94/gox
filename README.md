# gox简介

gox 是一个由 Go 语言（golang）编写的网络库。
gox 的关注点：
* 模块组合机制
* 可拆卸分布式，通过组装不同的模块，可随时随意把模块拆出来作为独立的服务器运行
* 支持服务注册与发现，可随时随意获取最新的服务器列表
* 支持TCP、WebSocket、KCP。
* 支持Protobuf、Json、SProto数据格式
* 通过grpc或内置rpcx系统，轻松搞定跨服务间的通信。
* 支持Actor,任何对象通过组合location.Location，且实现了LocationID()，都可以进行定位注册，之后不管这个对象在哪个服务器，都可以通过LocationID()直接发送消息


# API的简介 go版本（1.21.6）

现在让我们来看看如果创建一个服务器：
```
	var appConfPath string
	flag.StringVar(&appConfPath, "appConf", "app_1.yaml", "启动配置")
	flag.Parse()
	if appConfPath == "" {
		log.Fatalf("需要启动配置文件路径")
	}
	gox.Init(appConfPath)//初始化
	network := network.New() //创建网络系统
	network.SetInteriorService(new(kcp.KService), codechelper.Json) //设置内部通信服务类型和解析方式
	network.SetOutsideService(new(ws.WService), codechelper.Json)//设置外部通信服务类型和解析方式
	gox.SetNetWork(network)//设置网络系统
	gox.SetModule(new(mods.MainModule))//设置启动模块
	gox.Run()
  
```
主模块：
```
type (
//MainModule 主模块
MainModule struct {
	gox.Module//必须组合次Module
}
)

func (m *MainModule) OnInit() {
  //通过服务类型组装不同的模块
	switch gox.Config.AppType {
	case game.Gate:
		m.Put(&gate.GateModule{})		
	case game.Login:
		m.Put(&login.LoginModule{})		
	case game.Scene:
		m.Put(&scene.SceneModule{})		
	default:
		m.Put(&gate.GateModule{})
		m.Put(&login.LoginModule{})
		m.Put(&scene.SceneModule{})		
	}
}
```
 如何接受消息：
```
type (
	//LoginModule 登录模块
	LoginModule struct {
		gox.Module
	}
)

//OnInit 初始化
func (m *LoginModule) OnInit() {//协议注册得在初始化方法里进行
        //进行登录请求的注册，此方法时RPC放回，需客户端约束为RPC请求
	protoreg.RegisterRpcCmd(pb.CMD_C2S_LoginGame, m.LoginGame)
        //如果客户端不是RPC请求，则可以使用protoreg.Register,区别在于，你需要通过Session.Send返回客户端
        protoreg.Register(pb.CMD_C2S_LoginGame, m.LoginGame_1)
}
func (m *LoginModule) LoginGame(ctx context.Context, session types.ISession, req *pb.C2S_LoginGame) (*pb.S2C_LoginGame, error) {

	cfgs := gox.NetWork.GetServiceEntitys(types.WithType(game.Gate)) //获取Gate服务器配置
	if len(cfgs) == 0 {
		logger.Error().Msgf("没获取到[%s]对应的服务器配置", game.Gate)
		return &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown}, nil
	}
	gateCfg := cfgs[0]
	logger.Info().Msgf("[Rpcaddr:%s]", gateCfg.GetRpcAddr())
	conn := gox.NetWork.Rpc().GetClientConnByAddr(gateCfg.GetRpcAddr()) //创建session连接Gate服务器
	loginSession := pb.NewILoginGameClient(conn)
	resp, err := loginSession.LoginGame(ctx, req) //向Gate服务器请求token
	if err != nil {
		logger.Debug().Err(err)
		return &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown}, nil
	}
	return resp, nil //结果返回客户端
}

func (m *LoginModule) LoginGame_1(ctx context.Context, session types.ISession, req *pb.C2S_LoginGame) {

	cfgs := gox.NetWork.GetServiceEntitys(types.WithType(game.Gate)) //获取Gate服务器配置
	if len(cfgs) == 0 {
		logger.Error().Msgf("没获取到[%s]对应的服务器配置", game.Gate)
		session.Send(pb.CMD_C2S_LoginGame, &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown})
		return
	}
	gateCfg := cfgs[0]
	logger.Info().Msgf("[Rpcaddr:%s]", gateCfg.GetRpcAddr())
	conn := gox.NetWork.Rpc().GetClientConnByAddr(gateCfg.GetRpcAddr()) //创建session连接Gate服务器
	loginSession := pb.NewILoginGameClient(conn)
	resp, err := loginSession.LoginGame(ctx, req) //向Gate服务器请求token
	if err != nil {
		logger.Debug().Err(err)
		session.Send(pb.CMD_C2S_LoginGame, &pb.S2C_LoginGame{Error: pb.ErrCode_UnKnown})
		return
	}
	session.Send(pb.CMD_C2S_LoginGame, resp)
}

```

Loc注册和发送
```  
//注册
type (
	Scene struct {
		location.Location
		Id    uint
	}
)
func newScene(id uint) *Scene {	
  scene := &Scene{Id: id}
        gox.Location.Register(scene) //把场景添加进Location
	return scene
}
func (s *Scene) OnInit() {
	protoreg.AddLocationRpc(s, s.OnEnterScene)//添加消息回调
}

//所有Location对象都得实现此方法
func (s *Scene) LocationID() uint32 {
	return uint32(s.Id)
}

func (s *Scene) OnEnterScene(ctx context.Context, req *netpack.L2S_Enter) *netpack.S2L_Enter {
	return &netpack.S2L_Enter{Code: 0}
}
//发送
gox.Location.Send(locationID, &netpack.L2S_Enter{UnitId: req.UnitId}) 
//通过rpcx请求 b:bool值
backRsp := &netpack.S2L_Enter{}
b := gox.Location.Call(locationID, &netpack.L2S_Enter{UnitId: req.UnitId}, backRsp).Await() 
```

# examples运行
```
git clone https://github.com/xhaoh94/gox
```
下载etcd服务:[https://github.com/coreos/etcd/releases](https://github.com/coreos/etcd/releases)
```
1、终端执行 go mod init github.com/xhaoh94/gox ，生成go.mod后，在go.mod文件写上下面代码
replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/bbolt => github.com/coreos/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
2、终端执行 go mod tidy，等待拉取代码完毕(如果存在墙的问题，请提前设置好GOPROXY为https://goproxy.cn，具体步骤可以百度)
3、启动etcd服务
4、打开examples/sv/的终端 执行 go run main.go -sid 1 -type gate -iAddr 127.0.0.1:10001 -oAddr 127.0.0.1:10002
  再打开一个examples/sv/的终端 执行 go run main.go -sid 2 -type login -iAddr 127.0.0.1:20001 -oAddr 127.0.0.1:20002
  再打开一个examples/sv/的终端 执行 go run main.go -sid 3 -type scene -iAddr 127.0.0.1:30001 -oAddr 127.0.0.1:30002
  如果一些顺利的话，以上启动了3个服务器
  然后打开examples/cl/的终端(模拟客户端行为) 执行 go run main.go
  没问题的话，就可以看到打印的日志啦！
```


