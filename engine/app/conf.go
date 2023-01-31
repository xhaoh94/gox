package app

import (
	"io/ioutil"
	"log"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type (
	appConf struct {
		Log       logConf       `yaml:"log"`
		MongoDb   mongoDbConf   `yaml:"mongodb"`
		Network   networkConf   `yaml:"network"`
		WebSocket webSocketConf `yaml:"webSocket"`
		Etcd      etcdConf      `yaml:"etcd"`
	}
	logConf struct {
		LogPath     string `yaml:"log_path"`
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
	mongoDbConf struct {
		Url      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	}
	networkConf struct {
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
	webSocketConf struct {
		WebSocketMessageType int    `yaml:"ws_message_type"`
		WebSocketPattern     string `yaml:"ws_pattern"`
		WebSocketPath        string `yaml:"ws_path"`
		WebSocketScheme      string `yaml:"ws_scheme"`
		CertFile             string `yaml:"ws_certfile"`
		KeyFile              string `yaml:"ws_keyfile"`
	}
	etcdConf struct {
		EtcdList      []string      `yaml:"etcd_list"`
		EtcdTimeout   time.Duration `yaml:"etcd_timeout"`
		EtcdLeaseTime int64         `yaml:"etcd_lease_time"`
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
			Console:     "console",
			Skip:        2,
			Split:       false,
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
		},
		WebSocket: webSocketConf{
			WebSocketMessageType: 2,
			WebSocketPattern:     "ws",
			WebSocketPath:        "/ws",
			WebSocketScheme:      "/ws",
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
	bytes, err := ioutil.ReadFile(appConfPath)
	if err != nil {
		log.Printf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return
	}
	err = yaml.Unmarshal(bytes, &appCfg)
	if err != nil {
		log.Printf("LoadAppConfig err:[%v] path:[%s]", err, appConfPath)
		return
	}
	appCfg.Network.ReConnectInterval *= time.Second
	appCfg.Network.Heartbeat *= time.Second
	appCfg.Network.ConnectTimeout *= time.Second
	appCfg.Network.ReadTimeout *= time.Second
	appCfg.Etcd.EtcdTimeout *= time.Second
}
