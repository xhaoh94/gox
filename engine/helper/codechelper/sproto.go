package codechelper

import sproto "github.com/xjdrew/gosproto"

type sprotoCodec struct{}

var Sproto sprotoCodec

func (sprotoCodec) Encode(msg interface{}) (bytes []byte, err error) {
	bytes, err = sproto.Encode(msg)
	return
}
func (sprotoCodec) Decode(bytes []byte, msg interface{}) (err error) {
	_, err = sproto.Decode(bytes, msg)
	return
}
