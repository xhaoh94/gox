package types

type (
	IByteArray interface {
		AppendByte(byte)
		AppendBytes([]byte)
		AppendUint16(uint16)
		AppendInt16(int16)
		AppendUint32(uint32)
		AppendInt32(int32)
		AppendUint64(uint64)
		AppendInt64(int64)
		AppendString(string)
		AppendMessage(interface{}, ICodec) error
		ReadOneByte() byte
		ReadBytes(uint32) []byte
		ReadUint16() uint16
		ReadInt16() int16
		ReadUint32() uint32
		ReadInt32() int32
		ReadUint64() uint64
		ReadInt64() int64
		ReadString() string
		ReadMessage(interface{}, ICodec) error
		Position() uint32
		Length() uint32
		RemainLength() uint32
		Data() []byte
		PktData() []byte
		Release()
	}
)
