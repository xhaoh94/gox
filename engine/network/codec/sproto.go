package codec

import sproto "github.com/xjdrew/gosproto"

type sprotoCodec struct{}

var Sproto sprotoCodec

func (sprotoCodec) Marshal(msg interface{}) (data []byte, err error) {
	data, err = sproto.Encode(msg)
	return
}
func (sprotoCodec) Unmarshal(data []byte, msg interface{}) (err error) {
	_, err = sproto.Decode(data, msg)
	return
}
