package netpack

const (
	CMD_C2G_Login uint32 = 1000
	CMD_G2C_Login uint32 = 1000
	CMD_C2L_Login uint32 = 1001
	CMD_L2C_Login uint32 = 1001
	CMD_C2L_Enter uint32 = 1002
	CMD_L2C_Enter uint32 = 1002

	CMD_S2L_SayHello uint32 = 1003
)

type (
	C2G_Login struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	G2C_Login struct {
		Code  uint   `json:"code"`
		Addr  string `json:"addr"`
		Token string `json:"token"`
	}

	C2L_Login struct {
		User  string `json:"user"`
		Token string `json:"token"`
	}
	L2C_Login struct {
		Code uint `json:"code"`
	}

	C2L_Enter struct {
		SceneId uint `json:"sceneid"`
	}
	L2C_Enter struct {
		Code uint `json:"code"`
	}

	G2L_Login struct {
		User string `json:"user"`
	}
	L2G_Login struct {
		Token string `json:"token"`
	}

	L2S_Enter struct {
		UnitId uint `json:"unitid"`
	}
	S2L_Enter struct {
		Code uint `json:"code"`
	}
	L2S_SayHello struct {
		Txt string `json:"txt"`
	}
	S2L_SayHello struct {
		BackTxt string `json:"backtxt"`
	}
)
