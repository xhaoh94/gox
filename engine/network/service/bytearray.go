package service

import (
	"encoding/binary"
	"sync"

	"github.com/xhaoh94/gox/engine/consts"
	"github.com/xhaoh94/gox/engine/helper/codechelper"
	"github.com/xhaoh94/gox/engine/types"
)

// |------------------------------------------|
// msglen 包的总长度
// key    预留8位字节
// type   包数据类型(0x01:单向请求 0x02:心跳请求 0x03:心跳响应 0x04:rpc请求 0x05:rpc响应)
// cmd    数据结构对应的cmd
// rpc    rpc请求或响应时附带的rpcid
// msg    最终包体数据
// |-------------------------------------------------|
// [ 必填 ] [ 必填  ]  [ 必填 ]  [ 必填  ] [ 选填 ] [  选填  ]
// [msglen] [  key  ] [ type ]  [ cmd  ]  [  rpc ] [  msg  ]
// [uint16] [[8]byte] [[1]byte] [uint32]  [uint32] [[n]byte]
// |------------------------------------------------|

var bytePool sync.Pool = sync.Pool{New: func() any { return &ByteArray{} }}

// ByteArray 默认包体格式
type ByteArray struct {
	position uint32
	data     []byte
	endian   binary.ByteOrder
}

func NewByteArray(data []byte, endian binary.ByteOrder) *ByteArray {
	bytearray := bytePool.Get().(*ByteArray)
	bytearray.data = data
	bytearray.position = 0
	bytearray.endian = endian
	return bytearray
}
func (bytearray *ByteArray) AppendByte(v byte) {
	bytearray.data = append(bytearray.data, v)
}
func (bytearray *ByteArray) AppendBytes(v []byte) {
	bytearray.data = append(bytearray.data, v...)
}
func (bytearray *ByteArray) AppendUint16(v uint16) {
	bytes := codechelper.Uint16ToBytes(v, bytearray.endian)
	bytearray.data = append(bytearray.data, bytes...)
}
func (bytearray *ByteArray) AppendInt16(v int16) {
	bytes := codechelper.Int16ToBytes(v, bytearray.endian)
	bytearray.data = append(bytearray.data, bytes...)
}
func (bytearray *ByteArray) AppendUint32(v uint32) {
	bytes := codechelper.Uint32ToBytes(v, bytearray.endian)
	bytearray.data = append(bytearray.data, bytes...)
}
func (bytearray *ByteArray) AppendInt32(v int32) {
	bytes := codechelper.Int32ToBytes(v, bytearray.endian)
	bytearray.data = append(bytearray.data, bytes...)
}
func (bytearray *ByteArray) AppendUint64(v uint64) {
	bytes := codechelper.Uint64ToBytes(v, bytearray.endian)
	bytearray.data = append(bytearray.data, bytes...)
}
func (bytearray *ByteArray) AppendInt64(v int64) {
	bytes := codechelper.Int64ToBytes(v, bytearray.endian)
	bytearray.data = append(bytearray.data, bytes...)
}
func (bytearray *ByteArray) AppendString(v string) {
	l := len(v)
	if l <= 0 {
		return
	}
	bytearray.AppendUint16(uint16(l))
	bytearray.data = append(bytearray.data, []byte(v)...)
}
func (bytearray *ByteArray) AppendMessage(msg any, codec types.ICodec) error {
	if msg == nil {
		return consts.Error_2
	}
	msgData, err := codec.Marshal(msg)
	if err != nil {
		return err
	}
	bytearray.AppendBytes(msgData)
	return nil
}
func (bytearray *ByteArray) ReadOneByte() byte {
	bytes := bytearray.data[bytearray.position]
	bytearray.position++
	return bytes
}
func (bytearray *ByteArray) ReadBytes(size uint32) []byte {
	bytes := bytearray.data[bytearray.position : bytearray.position+size]
	bytearray.position += size
	return bytes
}

func (bytearray *ByteArray) ReadUint16() uint16 {
	r := codechelper.BytesToUint16(bytearray.data[bytearray.position:bytearray.position+2], bytearray.endian)
	bytearray.position += 2
	return r
}

func (bytearray *ByteArray) ReadInt16() int16 {
	r := codechelper.BytesToint16(bytearray.data[bytearray.position:bytearray.position+2], bytearray.endian)
	bytearray.position += 2
	return r
}

func (bytearray *ByteArray) ReadUint32() uint32 {
	r := codechelper.BytesToUint32(bytearray.data[bytearray.position:bytearray.position+4], bytearray.endian)
	bytearray.position += 4
	return r
}

func (bytearray *ByteArray) ReadInt32() int32 {
	r := codechelper.BytesToint32(bytearray.data[bytearray.position:bytearray.position+4], bytearray.endian)
	bytearray.position += 4
	return r
}

func (bytearray *ByteArray) ReadUint64() uint64 {
	r := codechelper.BytesToUint64(bytearray.data[bytearray.position:bytearray.position+8], bytearray.endian)
	bytearray.position += 8
	return r
}

func (bytearray *ByteArray) ReadInt64() int64 {
	r := codechelper.BytesToint64(bytearray.data[bytearray.position:bytearray.position+8], bytearray.endian)
	bytearray.position += 8
	return r
}
func (bytearray *ByteArray) ReadString() string {
	l := bytearray.ReadUint16()
	bytes := bytearray.ReadBytes(uint32(l))
	return string(bytes)
}

func (bytearray *ByteArray) ReadMessage(msg any, cdc types.ICodec) error {
	msgData := bytearray.RemainData()
	if err := cdc.Unmarshal(msgData, msg); err != nil {
		return err
	}
	return nil
}

func (bytearray *ByteArray) Position() uint32 {
	return bytearray.position
}

func (bytearray *ByteArray) Length() uint32 {
	return uint32(len(bytearray.data))
}
func (bytearray *ByteArray) RemainLength() uint32 {
	return bytearray.Length() - bytearray.Position()
}
func (bytearray *ByteArray) RemainData() []byte {
	return bytearray.ReadBytes(bytearray.RemainLength())
}

func (bytearray *ByteArray) Data() []byte {
	bytes := codechelper.Uint16ToBytes(uint16(len(bytearray.data)), bytearray.endian)
	bytes = append(bytes, bytearray.data...)
	return bytes
}

func (bytearray *ByteArray) Release() {
	bytearray.data = nil
	bytearray.position = 0
	bytearray.endian = nil
	bytePool.Put(bytearray)
}
