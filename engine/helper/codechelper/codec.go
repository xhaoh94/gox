package codechelper

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

type Number interface {
	int | uint | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func BytesTo[T Number](b []byte, endian binary.ByteOrder) T {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp T
	binary.Read(bytesBuffer, endian, &tmp)
	return tmp
}

func ToBytes[T Number](n T, endian binary.ByteOrder) []byte {
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
