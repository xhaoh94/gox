# gox简介
==================
gox 是一个由 Go 语言（golang）编写的网络库。适用于各类游戏服务器的开发。
gox 的关注点：
* 模块组合机制，模块内可通过api快捷方便的注册消息
* 可拆卸分布式，通过组装不同的模块，可随时随意把模块拆出来作为独立的服务器运行
* 支持服务注册与发现，可随时随意获取最新的服务器列表
* 支持TCP、WebSocket、KCP。
* 支持Protobuf、Json、SProto数据格式
* 通过GRpc或内置rpc系统，轻松搞定跨服务间的通信。
* 支持Actor,通过Actor注册，不管这个对象在哪个服务器，都可以通过ActorID直接发送消息
==================
# gox 获取和使用

获取：
```
git clone https://github.com/xhaoh94/gox
```
使用：
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
组合模块：
```
type (
	//MainModule 主模块
	MainModule struct {
		gox.Module//必须继承此
	}
)

func (m *MainModule) OnInit() {
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
例子运行：
```
1、终端执行 go mod init github.com/xhaoh94/gox ，生成go.mod后，在go.mod文件写上下面代码
replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.etcd.io/bbolt => github.com/coreos/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
2、终端执行 go mod tidy，等待拉取代码完毕(如果存在墙的问题，请提前设置好GOPROXY为https://goproxy.cn，具体步骤可以百度)
3、启动etcd服务，（如果没有，可以自行前往下载 https://github.com/coreos/etcd/releases）
4、打开examples/sv/的终端 执行 go run main.go -sid 1 -type gate -iAddr 127.0.0.1:10001 -oAddr 127.0.0.1:10002
  再打开一个examples/sv/的终端 执行 go run main.go -sid 2 -type login -iAddr 127.0.0.1:20001 -oAddr 127.0.0.1:20002
  再打开一个examples/sv/的终端 执行 go run main.go -sid 3 -type scene -iAddr 127.0.0.1:30001
  如果一些顺利的话，以上启动了3个服务器
  然后打开examples/cl/的终端(模拟客户端行为) 执行 go run main.go
  没问题的话，就可以看到打印的日志啦！
```


