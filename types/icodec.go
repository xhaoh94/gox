package types

type (
	//ICodec 解码编码接口
	ICodec interface {
		Encode(interface{}) ([]byte, error)
		Decode([]byte, interface{}) error
	}
)
