package codec

import "github.com/gogo/protobuf/proto"

type protobufCodec struct{}

var Protobuf protobufCodec

func (protobufCodec) Marshal(msg interface{}) (data []byte, err error) {
	data, err = proto.Marshal(msg.(proto.Message))
	return
}
func (protobufCodec) Unmarshal(data []byte, msg interface{}) (err error) {
	err = proto.Unmarshal(data, msg.(proto.Message))
	return
}
