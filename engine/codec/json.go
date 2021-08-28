package codec

import "encoding/json"

type jsonCodec struct{}

var Json jsonCodec

func (jsonCodec) Encode(msg interface{}) (bytes []byte, err error) {
	bytes, err = json.Marshal(msg)
	return
}
func (jsonCodec) Decode(bytes []byte, msg interface{}) (err error) {
	err = json.Unmarshal(bytes, msg)
	return
}
