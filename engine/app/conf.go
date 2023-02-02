package app

import (
	"time"
)

type (
	AppConf struct {
		Eid          uint          `yaml:"eid"`
		EType        string        `yaml:"etype"`
		Version      string        `yaml:"version"`
		InteriorAddr string        `yaml:"interioraddr"`
		OutsideAddr  string        `yaml:"outsideaddr"`
		RpcAddr      string        `yaml:"rpcaddr"`
		Log          LogConf       `yaml:"log"`
		Db           DbConf        `yaml:"db"`
		Network      NetworkConf   `yaml:"network"`
		WebSocket    WebSocketConf `yaml:"webSocket"`
		Etcd         EtcdConf      `yaml:"etcd"`
	}
	LogConf struct {
		Filename    string `yaml:"log_file_name"`
		IsWriteLog  bool   `yaml:"log_write_open"`
		Stacktrace  string `yaml:"log_stacktrace"`
		LogLevel    string `yaml:"log_level"`
		LogMaxSize  int    `yaml:"log_max_size"`
		MaxBackups  int    `yaml:"log_max_backups"`
		LogMaxAge   int    `yaml:"log_max_age"`
		Development bool   `yaml:"log_development"`
		Console     string `yaml:"log_console"`
		Skip        int    `yaml:"log_callerskip"`
		Split       bool   `yaml:"log_split_level"`
	}
	DbConf struct {
		Url      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	}
	NetworkConf struct {
		//SendMsgMaxLen 发送最大长度(websocket的话不能超过126) 默认0 不分片
		SendMsgMaxLen int `yaml:"send_msg_max_len"`
		//ReadMsgMaxLen 包体最大长度
		ReadMsgMaxLen int `yaml:"read_msg_max_len"`
		//ReConnectInterval 链接间隔
		ReConnectInterval time.Duration `yaml:"reconnect_interval"`
		//ConnectMax 尝试链接最大次数
		ReConnectMax int `yaml:"reconnection_max"`
		//Heartbeat 心跳时间
		Heartbeat time.Duration `yaml:"heartbeat"`
		//ConnectTimeout 链接超时
		ConnectTimeout time.Duration `yaml:"connect_timeout"`
		//ReadTimeout 读超时
		ReadTimeout time.Duration `yaml:"read_timeout"`
	}
	WebSocketConf struct {
		WebSocketMessageType int    `yaml:"ws_message_type"`
		WebSocketPattern     string `yaml:"ws_pattern"`
		WebSocketPath        string `yaml:"ws_path"`
		WebSocketScheme      string `yaml:"ws_scheme"`
		CertFile             string `yaml:"ws_certfile"`
		KeyFile              string `yaml:"ws_keyfile"`
	}
	EtcdConf struct {
		EtcdList      []string      `yaml:"etcd_list"`
		EtcdTimeout   time.Duration `yaml:"etcd_timeout"`
		EtcdLeaseTime int64         `yaml:"etcd_lease_time"`
	}
)

// func initCfg() {
// 	AppCfg = &AppConf{
// 		Log: LogConf{
// 			LogPath:     "./log/",
// 			IsWriteLog:  false,
// 			Stacktrace:  "error",
// 			LogLevel:    "debug",
// 			LogMaxSize:  128,
// 			MaxBackups:  30,
// 			LogMaxAge:   7,
// 			Development: true,
// 			Console:     "console",
// 			Skip:        2,
// 			Split:       false,
// 		},
// 		MongoDb: MongoDbConf{
// 			Url:      "127.0.0.1:27017",
// 			User:     "",
// 			Password: "",
// 			Database: "gox",
// 		},
// 		Network: NetworkConf{
// 			SendMsgMaxLen:     0,
// 			ReadMsgMaxLen:     2048,
// 			ReConnectInterval: 3 * time.Second,
// 			Heartbeat:         30 * time.Second,
// 			ConnectTimeout:    3 * time.Second,
// 			ReadTimeout:       35 * time.Second,
// 		},
// 		WebSocket: WebSocketConf{
// 			WebSocketMessageType: 2,
// 			WebSocketPattern:     "ws",
// 			WebSocketPath:        "/ws",
// 			WebSocketScheme:      "/ws",
// 			CertFile:             "",
// 			KeyFile:              "",
// 		},
// 		Etcd: EtcdConf{
// 			EtcdList:      []string{"127.0.0.1:2379"},
// 			EtcdTimeout:   5 * time.Second,
// 			EtcdLeaseTime: 5,
// 		},
// 	}
// }
