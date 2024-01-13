package gox

import (
	"encoding/binary"
	"time"
)

type (
	AppConf struct {
		Development  bool          `yaml:"development"`
		AppID        uint          `yaml:"app_id"`
		AppType      string        `yaml:"app_type"`
		Version      string        `yaml:"version"`
		InteriorAddr string        `yaml:"interioraddr"`
		OutsideAddr  string        `yaml:"outsideaddr"`
		RpcAddr      string        `yaml:"rpcaddr"`
		Location     bool          `yaml:"location"`
		LogConfPath  string        `yaml:"log_config_path"`
		Db           DbConf        `yaml:"db"`
		Network      NetworkConf   `yaml:"network"`
		WebSocket    WebSocketConf `yaml:"webSocket"`
		Etcd         EtcdConf      `yaml:"etcd"`
	}
	DbConf struct {
		Url      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	}
	NetworkConf struct {
		Endian binary.ByteOrder `yaml:"endian"`
		//发送最大长度(websocket的话不能超过126) 默认0 不分片
		SendMsgMaxLen int `yaml:"send_msg_max_len"`
		//包体最大长度
		ReadMsgMaxLen int `yaml:"read_msg_max_len"`
		//链接间隔
		ReConnectInterval time.Duration `yaml:"reconnect_interval"`
		//尝试链接最大次数
		ReConnectMax int `yaml:"reconnection_max"`
		//心跳时间
		Heartbeat time.Duration `yaml:"heartbeat"`
		//链接超时
		ConnectTimeout time.Duration `yaml:"connect_timeout"`
		//读超时
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

func (ut *NetworkConf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		Endian string `yaml:"endian"`
		//发送最大长度(websocket的话不能超过126) 默认0 不分片
		SendMsgMaxLen int `yaml:"send_msg_max_len"`
		//包体最大长度
		ReadMsgMaxLen int `yaml:"read_msg_max_len"`
		//链接间隔
		ReConnectInterval int `yaml:"reconnect_interval"`
		//尝试链接最大次数
		ReConnectMax int `yaml:"reconnection_max"`
		//心跳时间
		Heartbeat int `yaml:"heartbeat"`
		//链接超时
		ConnectTimeout int `yaml:"connect_timeout"`
		//ReadTimeout 读超时
		ReadTimeout int `yaml:"read_timeout"`
	}

	var tmp alias
	if err := unmarshal(&tmp); err != nil {
		return err
	}
	if tmp.Endian == "littleEndian" {
		ut.Endian = binary.LittleEndian
	} else if tmp.Endian == "bigEndian" {
		ut.Endian = binary.BigEndian
	} else {
		ut.Endian = binary.LittleEndian
	}
	ut.SendMsgMaxLen = tmp.SendMsgMaxLen
	ut.ReadMsgMaxLen = tmp.ReadMsgMaxLen
	ut.ReConnectInterval = time.Duration(tmp.ReConnectInterval) * time.Second
	ut.ReConnectMax = tmp.ReConnectMax

	if tmp.Heartbeat > 0 {
		ut.Heartbeat = time.Duration(tmp.Heartbeat) * time.Second
	} else {
		ut.Heartbeat = 30 * time.Second
	}

	if tmp.ConnectTimeout > 0 {
		ut.ConnectTimeout = time.Duration(tmp.ConnectTimeout) * time.Second
	} else {
		ut.ConnectTimeout = 3 * time.Second
	}

	ut.ReadTimeout = time.Duration(tmp.ReadTimeout) * time.Second
	return nil
}

func (ut *EtcdConf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		EtcdList      []string `yaml:"etcd_list"`
		EtcdTimeout   int      `yaml:"etcd_timeout"`
		EtcdLeaseTime int64    `yaml:"etcd_lease_time"`
	}

	var tmp alias
	if err := unmarshal(&tmp); err != nil {
		return err
	}
	ut.EtcdList = tmp.EtcdList

	if tmp.EtcdTimeout > 0 {
		ut.EtcdTimeout = time.Duration(tmp.EtcdTimeout) * time.Second
	} else {
		ut.EtcdTimeout = 3 * time.Second
	}

	ut.EtcdLeaseTime = tmp.EtcdLeaseTime
	return nil
}
