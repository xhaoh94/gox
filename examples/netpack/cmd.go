package netpack

const (
	C2S_TEST uint32 = 1000
	S2S_TEST uint32 = 2000
)

type (
	ReqTest struct {
		Acc string `json:acc`
		Pwd string `json:pwd`
	}
	RspTest struct {
		Token string `json:token`
	}
)
