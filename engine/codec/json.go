package codec

import (
	"encoding/json"
)

type (
	JsonCodec struct {
	}
)

func (*JsonCodec) Encode(msg interface{}) (bytes []byte, err error) {
	bytes, err = json.Marshal(msg)
	return
}
func (*JsonCodec) Decode(bytes []byte, msg interface{}) (err error) {
	err = json.Unmarshal(bytes, msg)
	return
}
