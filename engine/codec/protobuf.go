package codec

import "github.com/golang/protobuf/proto"

type (
	ProtobufCodec struct {
	}
)

func (*ProtobufCodec) Encode(msg interface{}) (bytes []byte, err error) {
	bytes, err = proto.Marshal(msg.(proto.Message))
	return
}
func (*ProtobufCodec) Decode(bytes []byte, msg interface{}) (err error) {
	err = proto.Unmarshal(bytes, msg.(proto.Message))
	return
}
