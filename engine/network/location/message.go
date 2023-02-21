package location

type (
	LocationGetRequire struct {
		IDs []uint32
	}
	LocationGetResponse struct {
		Datas []LocationData
	}

	LocationForwardRequire struct {
		CMD     uint32
		Require []byte
		IsCall  bool
	}
	LocationForwardResponse struct {
		IsSuc    bool
		Response []byte
	}

	LocationAddRequire struct {
		Datas []LocationData
	}
	LocationAddResponse struct {
	}

	LocationRemoveRequire struct {
		IDs []uint32
	}
	LocationRemoveResponse struct {
	}

	LocationLockRequire struct {
		Lock bool
	}
	LocationLockResponse struct {
	}

	LocationData struct {
		LocationID uint32
		AppID      uint
	}
)
