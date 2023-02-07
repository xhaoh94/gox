package types

type (
	//解码编码接口
	ICodec interface {
		Marshal(interface{}) ([]byte, error)
		Unmarshal([]byte, interface{}) error
	}
)
