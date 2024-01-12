package location

const (
	LocationGet      uint32 = 220129
	LocationRelay    uint32 = 220306
	LocationRegister uint32 = 230424
)

func IsLocationCMD(cmd uint32) bool {
	switch cmd {
	case LocationGet:
		return true
	case LocationRelay:
		return true
	case LocationRegister:
		return true
	}
	return false
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
