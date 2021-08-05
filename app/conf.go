package app

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

type (
	appConf struct {
		Log       logConf       `ini:"log"`
		MongoDb   mongoDbConf   `ini:"mongodb"`
		Network   networkConf   `ini:"network"`
		WebSocket webSocketConf `ini:"webSocket"`
		Etcd      etcdConf      `ini:"etcd"`
	}
	logConf struct {
		LogPath     string `ini:"log_path"`
		IsWriteLog  bool   `ini:"log_write_open"`
		Stacktrace  string `ini:"log_stacktrace"`
		LogLevel    string `ini:"log_level"`
		LogMaxSize  int    `ini:"log_max_size"`
		MaxBackups  int    `ini:"log_max_backups"`
		LogMaxAge   int    `ini:"log_max_age"`
		Development bool   `ini:"log_development"`
	}
	mongoDbConf struct {
		Url      string `ini:"url"`
		User     string `ini:"user"`
		Password string `ini:"password"`
		Database string `ini:"database"`
	}
	networkConf struct {
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
	webSocketConf struct {
		WebSocketMessageType int    `ini:"ws_message_type"`
		WebSocketPattern     string `ini:"ws_pattern"`
		WebSocketPath        string `ini:"ws_path"`
		WebSocketScheme      string `ini:"ws_scheme"`
		CertFile             string `ini:"ws_certfile"`
		KeyFile              string `ini:"ws_keyfile"`
	}
	etcdConf struct {
		EtcdList      []string      `ini:"etcd_list"`
		EtcdTimeout   time.Duration `ini:"etcd_timeout"`
		EtcdLeaseTime int64         `ini:"etcd_lease_time"`
	}
)

var appCfg *appConf

func initCfg() {
	appCfg = &appConf{
		Log: logConf{
			LogPath:     "./log/",
			IsWriteLog:  false,
			Stacktrace:  "error",
			LogLevel:    "debug",
			LogMaxSize:  128,
			MaxBackups:  30,
			LogMaxAge:   7,
			Development: true,
		},
		MongoDb: mongoDbConf{
			Url:      "127.0.0.1:27017",
			User:     "",
			Password: "",
			Database: "gox",
		},
		Network: networkConf{
			SendMsgMaxLen:     0,
			ReadMsgMaxLen:     2048,
			ReConnectInterval: 3 * time.Second,
			Heartbeat:         30 * time.Second,
			ConnectTimeout:    3 * time.Second,
			ReadTimeout:       35 * time.Second,
			NetEndian:         "LittleEndian",
		},
		WebSocket: webSocketConf{
			WebSocketMessageType: 2,
			WebSocketPattern:     "ws",
			CertFile:             "",
			KeyFile:              "",
		},
		Etcd: etcdConf{
			EtcdList:      []string{"127.0.0.1:2379"},
			EtcdTimeout:   5 * time.Second,
			EtcdLeaseTime: 5,
		},
	}
}

func LoadAppConfig(appConfPath string) {
	appCfg = new(appConf)
	if err := ini.MapTo(appCfg, appConfPath); err != nil {
		log.Printf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return
	}

	appCfg.Network.ReConnectInterval *= time.Second
	appCfg.Network.Heartbeat *= time.Second
	appCfg.Network.ConnectTimeout *= time.Second
	appCfg.Network.ReadTimeout *= time.Second
	appCfg.Etcd.EtcdTimeout *= time.Second
}
