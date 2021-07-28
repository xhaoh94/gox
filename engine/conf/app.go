package conf

import (
	"encoding/binary"
	"log"
	"time"

	"github.com/xhaoh94/gox/app"

	"github.com/go-ini/ini"
)

type (
	AppConf struct {
		Version   string        `ini:"version"`
		Log       LogConf       `ini:"log"`
		MongoDb   MongoDbConf   `ini:"mongodb"`
		Network   NetworkConf   `ini:"network"`
		WebSocket WebSocketConf `ini:"webSocket"`
		Etcd      EtcdConf      `ini:"etcd"`
	}
	LogConf struct {
		LogPath     string `ini:"log_path"`
		IsWriteLog  bool   `ini:"log_write_open"`
		Stacktrace  string `ini:"log_stacktrace"`
		LogLevel    string `ini:"log_level"`
		LogMaxSize  int    `ini:"log_max_size"`
		MaxBackups  int    `ini:"log_max_backups"`
		LogMaxAge   int    `ini:"log_max_age"`
		Development bool   `ini:"log_development"`
	}
	MongoDbConf struct {
		Url      string `ini:"url"`
		User     string `ini:"user"`
		Password string `ini:"password"`
		Database string `ini:"database"`
	}
	NetworkConf struct {
		//SendMsgMaxLen 发送最大长度(websocket的话不能超过126) 默认0 不分片
		SendMsgMaxLen int `ini:"send_msg_max_len"`
		//ReadMsgMaxLen 包体最大长度
		ReadMsgMaxLen int `ini:"read_msg_max_len"`
		//ReConnectInterval 链接间隔
		ReConnectInterval time.Duration `ini:"reconnect_interval"`
		//ConnectMax 尝试链接最大次数
		ReConnectMax int `ini:"reconnection_max"`
		//Heartbeat 心跳时间
		Heartbeat time.Duration `ini:"heartbeat"`
		//ConnectTimeout 链接超时
		ConnectTimeout time.Duration `ini:"connect_timeout"`
		//ReadTimeout 读超时
		ReadTimeout time.Duration `ini:"read_timeout"`
		NetEndian   string        `ini:"net_endian"`
	}
	WebSocketConf struct {
		WebSocketMessageType int    `ini:"websocket_message_type"`
		WebSocketPattern     string `ini:"websocket_pattern"`
	}
	EtcdConf struct {
		EtcdList      []string      `ini:"etcd_list"`
		EtcdTimeout   time.Duration `ini:"etcd_timeout"`
		EtcdLeaseTime int64         `ini:"etcd_lease_time"`
	}
)

var AppCfg *AppConf

func LoadAppConfig(appConfPath string) {
	AppCfg = new(AppConf)
	if err := ini.MapTo(AppCfg, appConfPath); err != nil {
		log.Printf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return
	}
	switch AppCfg.Network.NetEndian {
	case "LittleEndian":
		app.NetEndian = binary.LittleEndian
		break
	case "BigEndian":
		app.NetEndian = binary.BigEndian
		break
	}
	app.WebSocketMessageType = AppCfg.WebSocket.WebSocketMessageType
	app.WebSocketPattern = AppCfg.WebSocket.WebSocketPattern
	app.SendMsgMaxLen = AppCfg.Network.SendMsgMaxLen
	app.ReadMsgMaxLen = AppCfg.Network.ReadMsgMaxLen
	app.ReConnectInterval = AppCfg.Network.ReConnectInterval * time.Second
	app.ReConnectMax = AppCfg.Network.ReConnectMax
	app.Heartbeat = AppCfg.Network.Heartbeat * time.Second
	app.ConnectTimeout = AppCfg.Network.ConnectTimeout * time.Second
	app.ReadTimeout = AppCfg.Network.ReadTimeout * time.Second
	app.EtcdList = AppCfg.Etcd.EtcdList
	app.EtcdTimeout = AppCfg.Etcd.EtcdTimeout * time.Second
	app.EtcdLeaseTime = AppCfg.Etcd.EtcdLeaseTime
}
