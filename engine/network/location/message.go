package location

type (
	LocationGetRequire struct {
		IDs []uint32
	}
	LocationGetResponse struct {
		Datas []LocationData
	}

	LocationRelayRequire struct {
		CMD     uint32
		Require []byte
		IsCall  bool
	}
	LocationRelayResponse struct {
		IsSuc    bool
		Response []byte
	}

	LocationData struct {
		LocationID uint32
		AppID      uint
	}
)
