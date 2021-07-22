package app

import "time"

var (
	EtcdList                    = []string{"127.0.0.1:2379"}
	EtcdTimeout   time.Duration = (5 * time.Second)
	EtcdLeaseTime int64         = 5
)
