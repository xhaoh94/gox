package codechelper

import "encoding/json"

type jsonCodec struct{}

var Json jsonCodec

func (jsonCodec) Marshal(msg interface{}) (bytes []byte, err error) {
	bytes, err = json.Marshal(msg)
	return
}
func (jsonCodec) Unmarshal(bytes []byte, msg interface{}) (err error) {
	err = json.Unmarshal(bytes, msg)
	return
}
