package sv

import (
	"encoding/binary"
	"sync"

	"github.com/xhaoh94/gox/consts"
	"github.com/xhaoh94/gox/engine/codec"
	"github.com/xhaoh94/gox/types"
)

// |------------------------------------------|
// msglen 包的总长度
// key    预留8位字节
// type   包数据类型(0x01:单向请求 0x02:心跳请求 0x03:心跳响应 0x04:rpc请求 0x05:rpc响应)
// cmd    数据结构对应的cmd
// rpc    rpc请求或响应时附带的rpcid
// msg    最终包体数据
// |-------------------------------------------------|
// [ 必填  ][ 必填  ][ 必填 ][ 必填  ][ 选填 ][  选填  ]
// [msglen][  key  ][ type ][ cmd  ][  rpc ][  msg  ]
// [uint16][[8]byte][ byte ][uint32][uint32][[n]byte]
// |------------------------------------------------|

var bytePool sync.Pool = sync.Pool{New: func() interface{} { return &ByteArray{} }}

//ByteArray 默认包体格式
type ByteArray struct {
	position uint32
	data     []byte
	endian   binary.ByteOrder
}

func NewByteArray(data []byte, endian binary.ByteOrder) *ByteArray {
	b := bytePool.Get().(*ByteArray)
	b.data = data
	b.position = 0
	b.endian = endian
	return b
}
func (b *ByteArray) AppendByte(v byte) {
	b.data = append(b.data, v)
}
func (b *ByteArray) AppendBytes(v []byte) {
	b.data = append(b.data, v...)
}
func (b *ByteArray) AppendUint16(v uint16) {
	bytes := codec.Uint16ToBytes(v, b.endian)
	b.data = append(b.data, bytes...)
}
func (b *ByteArray) AppendInt16(v int16) {
	bytes := codec.Int16ToBytes(v, b.endian)
	b.data = append(b.data, bytes...)
}
func (b *ByteArray) AppendUint32(v uint32) {
	bytes := codec.Uint32ToBytes(v, b.endian)
	b.data = append(b.data, bytes...)
}
func (b *ByteArray) AppendInt32(v int32) {
	bytes := codec.Int32ToBytes(v, b.endian)
	b.data = append(b.data, bytes...)
}
func (b *ByteArray) AppendUint64(v uint64) {
	bytes := codec.Uint64ToBytes(v, b.endian)
	b.data = append(b.data, bytes...)
}
func (b *ByteArray) AppendInt64(v int64) {
	bytes := codec.Int64ToBytes(v, b.endian)
	b.data = append(b.data, bytes...)
}
func (b *ByteArray) AppendString(v string) {
	l := len(v)
	if l <= 0 {
		return
	}
	b.AppendUint16(uint16(l))
	b.data = append(b.data, []byte(v)...)
}
func (b *ByteArray) AppendMessage(msg interface{}, cdc types.ICodec) error {
	if msg == nil {
		return consts.CodecError
	}
	msgData, err := cdc.Encode(msg)
	if err != nil {
		return err
	}
	b.AppendBytes(msgData)
	return nil
}
func (b *ByteArray) ReadOneByte() byte {
	bytes := b.data[b.position]
	b.position++
	return bytes
}
func (b *ByteArray) ReadBytes(size uint32) []byte {
	bytes := b.data[b.position : b.position+size]
	b.position += size
	return bytes
}

func (b *ByteArray) ReadUint16() uint16 {
	r := codec.BytesToUint16(b.data[b.position:b.position+2], b.endian)
	b.position += 2
	return r
}

func (b *ByteArray) ReadInt16() int16 {
	r := codec.BytesToint16(b.data[b.position:b.position+2], b.endian)
	b.position += 2
	return r
}

func (b *ByteArray) ReadUint32() uint32 {
	r := codec.BytesToUint32(b.data[b.position:b.position+4], b.endian)
	b.position += 4
	return r
}

func (b *ByteArray) ReadInt32() int32 {
	r := codec.BytesToint32(b.data[b.position:b.position+4], b.endian)
	b.position += 4
	return r
}

func (b *ByteArray) ReadUint64() uint64 {
	r := codec.BytesToUint64(b.data[b.position:b.position+8], b.endian)
	b.position += 8
	return r
}

func (b *ByteArray) ReadInt64() int64 {
	r := codec.BytesToint64(b.data[b.position:b.position+8], b.endian)
	b.position += 8
	return r
}
func (b *ByteArray) ReadString() string {
	l := b.ReadUint16()
	bytes := b.ReadBytes(uint32(l))
	return string(bytes)
}

func (b *ByteArray) ReadMessage(msg interface{}, cdc types.ICodec) error {
	msgLen := b.Length() - b.Position()
	msgData := b.ReadBytes(msgLen)
	if err := cdc.Decode(msgData, msg); err != nil {
		return err
	}
	return nil
}

func (b *ByteArray) Position() uint32 {
	return b.position
}

func (b *ByteArray) Length() uint32 {
	return uint32(len(b.data))
}
func (b *ByteArray) RemainLength() uint32 {
	return b.Length() - b.Position()
}

func (b *ByteArray) Data() []byte {
	return b.data
}
func (b *ByteArray) PktData() []byte {
	bytes := codec.Uint16ToBytes(uint16(b.Length()), b.endian)
	bytes = append(bytes, b.data...)
	return bytes
}

func (b *ByteArray) Release() {
	b.data = nil
	b.position = 0
	b.endian = nil
	bytePool.Put(b)
}
