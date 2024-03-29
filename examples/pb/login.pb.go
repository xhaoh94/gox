// Code generated by protoc-gen-go. DO NOT EDIT.
// source: login/login.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

//cs=100100001
type C2S_LoginGame struct {
	Account              string   `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Password             string   `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *C2S_LoginGame) Reset()         { *m = C2S_LoginGame{} }
func (m *C2S_LoginGame) String() string { return proto.CompactTextString(m) }
func (*C2S_LoginGame) ProtoMessage()    {}
func (*C2S_LoginGame) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{0}
}

func (m *C2S_LoginGame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_C2S_LoginGame.Unmarshal(m, b)
}
func (m *C2S_LoginGame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_C2S_LoginGame.Marshal(b, m, deterministic)
}
func (m *C2S_LoginGame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_C2S_LoginGame.Merge(m, src)
}
func (m *C2S_LoginGame) XXX_Size() int {
	return xxx_messageInfo_C2S_LoginGame.Size(m)
}
func (m *C2S_LoginGame) XXX_DiscardUnknown() {
	xxx_messageInfo_C2S_LoginGame.DiscardUnknown(m)
}

var xxx_messageInfo_C2S_LoginGame proto.InternalMessageInfo

func (m *C2S_LoginGame) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *C2S_LoginGame) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type S2C_LoginGame struct {
	Error                ErrCode  `protobuf:"varint,1,opt,name=error,proto3,enum=ErrCode" json:"error,omitempty"`
	Addr                 string   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
	Token                string   `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *S2C_LoginGame) Reset()         { *m = S2C_LoginGame{} }
func (m *S2C_LoginGame) String() string { return proto.CompactTextString(m) }
func (*S2C_LoginGame) ProtoMessage()    {}
func (*S2C_LoginGame) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{1}
}

func (m *S2C_LoginGame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_S2C_LoginGame.Unmarshal(m, b)
}
func (m *S2C_LoginGame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_S2C_LoginGame.Marshal(b, m, deterministic)
}
func (m *S2C_LoginGame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_S2C_LoginGame.Merge(m, src)
}
func (m *S2C_LoginGame) XXX_Size() int {
	return xxx_messageInfo_S2C_LoginGame.Size(m)
}
func (m *S2C_LoginGame) XXX_DiscardUnknown() {
	xxx_messageInfo_S2C_LoginGame.DiscardUnknown(m)
}

var xxx_messageInfo_S2C_LoginGame proto.InternalMessageInfo

func (m *S2C_LoginGame) GetError() ErrCode {
	if m != nil {
		return m.Error
	}
	return ErrCode_Success
}

func (m *S2C_LoginGame) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *S2C_LoginGame) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func init() {
	proto.RegisterType((*C2S_LoginGame)(nil), "C2S_LoginGame")
	proto.RegisterType((*S2C_LoginGame)(nil), "S2C_LoginGame")
}

func init() { proto.RegisterFile("login/login.proto", fileDescriptor_6fe61ab550dd3bc4) }

var fileDescriptor_6fe61ab550dd3bc4 = []byte{
	// 194 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xcc, 0xc9, 0x4f, 0xcf,
	0xcc, 0xd3, 0x07, 0x93, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x52, 0xc2, 0xc9, 0xf9, 0xb9, 0xb9,
	0xf9, 0x79, 0xfa, 0x10, 0x0a, 0x2a, 0x28, 0x02, 0x15, 0x4c, 0x2d, 0x2a, 0x4a, 0xce, 0x4f, 0x49,
	0x85, 0x88, 0x2a, 0xb9, 0x72, 0xf1, 0x3a, 0x1b, 0x05, 0xc7, 0xfb, 0x80, 0x74, 0xbb, 0x27, 0xe6,
	0xa6, 0x0a, 0x49, 0x70, 0xb1, 0x27, 0x26, 0x27, 0xe7, 0x97, 0xe6, 0x95, 0x48, 0x30, 0x2a, 0x30,
	0x6a, 0x70, 0x06, 0xc1, 0xb8, 0x42, 0x52, 0x5c, 0x1c, 0x05, 0x89, 0xc5, 0xc5, 0xe5, 0xf9, 0x45,
	0x29, 0x12, 0x4c, 0x60, 0x29, 0x38, 0x5f, 0x29, 0x92, 0x8b, 0x37, 0xd8, 0xc8, 0x19, 0xc9, 0x18,
	0x39, 0x2e, 0xd6, 0xd4, 0xa2, 0xa2, 0xfc, 0x22, 0xb0, 0x21, 0x7c, 0x46, 0x1c, 0x7a, 0xae, 0x45,
	0x45, 0xce, 0xf9, 0x29, 0xa9, 0x41, 0x10, 0x61, 0x21, 0x21, 0x2e, 0x96, 0xc4, 0x94, 0x94, 0x22,
	0xa8, 0x41, 0x60, 0xb6, 0x90, 0x08, 0x17, 0x6b, 0x49, 0x7e, 0x76, 0x6a, 0x9e, 0x04, 0x33, 0x58,
	0x10, 0xc2, 0x71, 0xe2, 0x8e, 0x62, 0xd1, 0xb3, 0x2e, 0x48, 0x5a, 0xc5, 0xc4, 0x14, 0x90, 0x94,
	0xc4, 0x06, 0x76, 0xb5, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xc4, 0xb8, 0x08, 0xbf, 0xf5, 0x00,
	0x00, 0x00,
}
