package location

type (
	LocationGetRequire struct {
		IDs []uint32
	}
	LocationGetResponse struct {
		Datas []LocationData
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
