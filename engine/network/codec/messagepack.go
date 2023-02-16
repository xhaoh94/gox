package codec

import (
	"github.com/vmihailenco/msgpack/v5"
)

type msgpackCodec struct{}

var MsgPack msgpackCodec

func (msgpackCodec) Marshal(msg interface{}) (data []byte, err error) {
	data, err = msgpack.Marshal(msg)
	return
}
func (msgpackCodec) Unmarshal(data []byte, msg interface{}) (err error) {
	err = msgpack.Unmarshal(data, msg)
	return
}
