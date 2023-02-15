package codec

import "encoding/json"

type jsonCodec struct{}

var Json jsonCodec

func (jsonCodec) Marshal(msg interface{}) (data []byte, err error) {
	data, err = json.Marshal(msg)
	return
}
func (jsonCodec) Unmarshal(data []byte, msg interface{}) (err error) {
	err = json.Unmarshal(data, msg)
	return
}
