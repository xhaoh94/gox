package codec

import "github.com/golang/protobuf/proto"

type protobufCodec struct{}

var Protobuf protobufCodec

func (protobufCodec) Encode(msg interface{}) (bytes []byte, err error) {
	bytes, err = proto.Marshal(msg.(proto.Message))
	return
}
func (protobufCodec) Decode(bytes []byte, msg interface{}) (err error) {
	err = proto.Unmarshal(bytes, msg.(proto.Message))
	return
}
