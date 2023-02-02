package codechelper

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

// BytesToUint16 转uint16
func BytesToUint16(b []byte, endian binary.ByteOrder) uint16 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint16
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

// BytesToint16 转int16
func BytesToint16(b []byte, endian binary.ByteOrder) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int16
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

// BytesToUint32 转uint32
func BytesToUint32(b []byte, endian binary.ByteOrder) uint32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint32
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

// BytesToint32 转int32
func BytesToint32(b []byte, endian binary.ByteOrder) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

// BytesToUint64 转uint64
func BytesToUint64(b []byte, endian binary.ByteOrder) uint64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint64
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

// BytesToint64 转int64
func BytesToint64(b []byte, endian binary.ByteOrder) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int64
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

// Uint16ToBytes 转bytes
func Uint16ToBytes(n uint16, endian binary.ByteOrder) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, endian, &n)
	return bytesBuffer.Bytes()
}

// Int16ToBytes 转bytes
func Int16ToBytes(n int16, endian binary.ByteOrder) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, endian, &n)
	return bytesBuffer.Bytes()
}

// Uint32ToBytes 转bytes
func Uint32ToBytes(n uint32, endian binary.ByteOrder) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, endian, &n)
	return bytesBuffer.Bytes()
}

// Int32ToBytes 转bytes
func Int32ToBytes(n int32, endian binary.ByteOrder) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, endian, &n)
	return bytesBuffer.Bytes()
}

// Uint64ToBytes 转bytes
func Uint64ToBytes(n uint64, endian binary.ByteOrder) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, endian, &n)
	return bytesBuffer.Bytes()
}

// Int64ToBytes 转bytes
func Int64ToBytes(n int64, endian binary.ByteOrder) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, endian, &n)
	return bytesBuffer.Bytes()
}

// CompressBytes 压缩字节
func CompressBytes(data []byte) ([]byte, error) {

	var buf bytes.Buffer

	writer := zlib.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()

	return buf.Bytes(), nil
}

// DecompressBytes 解压字节
func DecompressBytes(data []byte) ([]byte, error) {

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	return io.ReadAll(reader)
}
