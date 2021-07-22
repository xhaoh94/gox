package codec

import gosproto "github.com/xjdrew/gosproto"

type (
	SprotoCodec struct {
	}
)

func (*SprotoCodec) Encode(msg interface{}) (bytes []byte, err error) {
	bytes, err = gosproto.Encode(msg)
	return
}
func (*SprotoCodec) Decode(bytes []byte, msg interface{}) (err error) {
	_, err = gosproto.Decode(bytes, msg)
	return
}
