# gox简介

gox 是一个由 Go 语言（golang）编写的网络库。适用于各类游戏服务器的开发。
gox 的关注点：
* 模块组合机制，模块内可通过api快捷方便的注册消息
* 可拆卸分布式，通过组装不同的模块，可随时随意把模块拆出来作为独立的服务器运行
* 支持服务注册与发现，可随时随意获取最新的服务器列表
* 支持TCP、WebSocket、KCP。
* 支持Protobuf、Json、SProto数据格式
* 通过GRpc或内置rpc系统，轻松搞定跨服务间的通信。
* 支持Actor,任何对象通过继承actor.Actor，且实现了ActorID()，都可以进行Actor注册，之后不管这个对象在哪个服务器，都可以通过ActorID直接发送消息

# API的简介

现在让我们来看看如果创建一个服务器：
```
  engine := gox.NewEngine(sid, sType, "1.0.0")//实例化一个服务器 传入id，服务器类型，和版本
	game.Engine = engine //全局存储
	engine.SetModule(new(mods.MainModule)) //设置启动模块
	engine.SetCodec(codec.Json) //设置数据结构
	engine.SetInteriorService(new(tcp.TService), iAddr)//设置内部通信类型
	engine.SetOutsideService(new(kcp.KService), oAddr)//设置外部通信类型	
	engine.Start("gox.ini")//启动服务
  
```
主模块：
```
type (
	//MainModule 主模块
	MainModule struct {
		gox.Module//必须继承此
	}
)

func (m *MainModule) OnInit() {
  //通过服务类型组装不同的模块
	switch m.GetEngine().ServiceType() {
	case game.Gate:
		m.Put(&gate.GateModule{})
		break
	case game.Login:
		m.Put(&login.LoginModule{})
		break
	case game.Scene:
		m.Put(&scene.SceneModule{})
		break
	default:
		m.Put(&gate.GateModule{})
		m.Put(&login.LoginModule{})
		m.Put(&scene.SceneModule{})
		break
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
	m.RegisterRPC(m.RspToken)//注册rpc回调
	m.Register(netpack.CMD_C2L_Login, m.RspLogin)//注册协议	
}
func (m *LoginModule) RspToken(ctx context.Context, req *netpack.G2L_Login) *netpack.L2G_Login { return &netpack.L2G_Login{Token: token} }
func (m *LoginModule) RspLogin(ctx context.Context, session types.ISession, req *netpack.C2L_Login){}
```
如果发送消息：
```
  cfgs := m.GetServiceConfListByType(game.Login) //获取login服务器配置
	loginCfg := cfgs[0]
	session := m.GetSessionByAddr(loginCfg.GetInteriorAddr()) //创建session连接login服务器
  
  //直接发送没有返回
	session.Send(netpack.CMD_C2G_Login, &netpack.C2G_Login{User: "xhaoh94", Password: "123456"})
  
  //rpc请求 b:bool值     
	Rsp_L2G_Login := &netpack.L2G_Login{}
	b := session.Call(&netpack.G2L_Login{User: msg.User}, Rsp_L2G_Login).Await()  
```
Actor注册和发送
```  
//注册
type (
	Scene struct {
		actor.Actor
		Id    uint
	}
)
func newScene(id uint) *Scene {	
  scene := &Scene{Id: id}
	scene.AddActorFn(s.OnUnitEnter) //添加Actor回调
	game.Engine.GetNetWork().GetActorCtrl().Add(scene) //把场景添加进Actor
	return scene
}

//ActorID 所有Acotr对象都得实现此方法
func (s *Scene) ActorID() uint32 {
	return uint32(s.Id)
}

func (s *Scene) OnUnitEnter(ctx context.Context, req *netpack.L2S_Enter) *netpack.S2L_Enter {
	return &netpack.S2L_Enter{Code: 0}
}
//发送
//直接发送没有返回
game.Engine.GetNetWork().GetActorCtrl().Send(actorId, &netpack.L2S_Enter{UnitId: req.UnitId}) 
//通过rpc请求 b:bool值
backRsp := &netpack.S2L_Enter{}
b := game.Engine.GetNetWork().GetActorCtrl().Call(actorId, &netpack.L2S_Enter{UnitId: req.UnitId}, backRsp).Await() 
```

# examples运行
```
git clone https://github.com/xhaoh94/gox
```
```
1、终端执行 go mod init github.com/xhaoh94/gox ，生成go.mod后，在go.mod文件写上下面代码
replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/bbolt => github.com/coreos/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
2、终端执行 go mod tidy，等待拉取代码完毕(如果存在墙的问题，请提前设置好GOPROXY为https://goproxy.cn，具体步骤可以百度)
3、启动etcd服务，如果没有下载，可以自行前往下载 [https://github.com/coreos/etcd/releases](etcd)
4、打开examples/sv/的终端 执行 go run main.go -sid 1 -type gate -iAddr 127.0.0.1:10001 -oAddr 127.0.0.1:10002
  再打开一个examples/sv/的终端 执行 go run main.go -sid 2 -type login -iAddr 127.0.0.1:20001 -oAddr 127.0.0.1:20002
  再打开一个examples/sv/的终端 执行 go run main.go -sid 3 -type scene -iAddr 127.0.0.1:30001
  如果一些顺利的话，以上启动了3个服务器
  然后打开examples/cl/的终端(模拟客户端行为) 执行 go run main.go
  没问题的话，就可以看到打印的日志啦！
```


