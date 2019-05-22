// Code generated by protoc-gen-go. DO NOT EDIT.
// source: server_test.proto

package rpc_proto

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

type Person struct {
	Name                 *string  `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Age                  *int32   `protobuf:"varint,2,req,name=age" json:"age,omitempty"`
	Email                *string  `protobuf:"bytes,3,req,name=email" json:"email,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Person) Reset()         { *m = Person{} }
func (m *Person) String() string { return proto.CompactTextString(m) }
func (*Person) ProtoMessage()    {}
func (*Person) Descriptor() ([]byte, []int) {
	return fileDescriptor_bcfed924a188e565, []int{0}
}

func (m *Person) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Person.Unmarshal(m, b)
}
func (m *Person) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Person.Marshal(b, m, deterministic)
}
func (m *Person) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Person.Merge(m, src)
}
func (m *Person) XXX_Size() int {
	return xxx_messageInfo_Person.Size(m)
}
func (m *Person) XXX_DiscardUnknown() {
	xxx_messageInfo_Person.DiscardUnknown(m)
}

var xxx_messageInfo_Person proto.InternalMessageInfo

func (m *Person) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Person) GetAge() int32 {
	if m != nil && m.Age != nil {
		return *m.Age
	}
	return 0
}

func (m *Person) GetEmail() string {
	if m != nil && m.Email != nil {
		return *m.Email
	}
	return ""
}

func init() {
	proto.RegisterType((*Person)(nil), "rpc_proto.Person")
}

func init() { proto.RegisterFile("server_test.proto", fileDescriptor_bcfed924a188e565) }

var fileDescriptor_bcfed924a188e565 = []byte{
	// 108 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2c, 0x4e, 0x2d, 0x2a,
	0x4b, 0x2d, 0x8a, 0x2f, 0x49, 0x2d, 0x2e, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2c,
	0x2a, 0x48, 0x8e, 0x07, 0x33, 0x95, 0x5c, 0xb8, 0xd8, 0x02, 0x52, 0x8b, 0x8a, 0xf3, 0xf3, 0x84,
	0x84, 0xb8, 0x58, 0xf2, 0x12, 0x73, 0x53, 0x25, 0x18, 0x15, 0x98, 0x34, 0x38, 0x83, 0xc0, 0x6c,
	0x21, 0x01, 0x2e, 0xe6, 0xc4, 0xf4, 0x54, 0x09, 0x26, 0x05, 0x26, 0x0d, 0xd6, 0x20, 0x10, 0x53,
	0x48, 0x84, 0x8b, 0x35, 0x35, 0x37, 0x31, 0x33, 0x47, 0x82, 0x19, 0xac, 0x0c, 0xc2, 0x01, 0x04,
	0x00, 0x00, 0xff, 0xff, 0x46, 0x4c, 0xbd, 0xf9, 0x64, 0x00, 0x00, 0x00,
}
