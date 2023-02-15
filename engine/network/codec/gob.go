package codec

import (
	"bytes"
	"encoding/gob"
)

type gobCodec struct{}

var Gob gobCodec

func (gobCodec) Marshal(msg interface{}) (data []byte, err error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(msg)
	if err != nil {
		return
	}
	data = buffer.Bytes()
	return
}
func (gobCodec) Unmarshal(data []byte, msg interface{}) (err error) {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(msg)
	return
}
