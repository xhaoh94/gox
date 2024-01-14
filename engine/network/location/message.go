package location

import "github.com/xhaoh94/gox/engine/helper/strhelper"

var (
	LocationGet      uint32
	LocationRelay    uint32
	LocationRegister uint32
)

func init() {
	LocationGet = strhelper.StringToHash("LocationGet")
	LocationRelay = strhelper.StringToHash("LocationRelay")
	LocationRegister = strhelper.StringToHash("LocationRegister")
}

type (
	LocationGetRequire struct {
		IDs []uint32
	}
	LocationGetResponse struct {
		Datas []LocationData
	}

	LocationRelayRequire struct {
		LocationID uint32
		CMD        uint32
		Require    []byte
		IsCall     bool
	}
	LocationRelayResponse struct {
		IsSuc    bool
		Response []byte
	}

	LocationRegisterRequire struct {
		IsRegister  bool
		AppID       uint
		LocationIDs []uint32
	}

	LocationData struct {
		LocationID uint32
		AppID      uint
	}
)
