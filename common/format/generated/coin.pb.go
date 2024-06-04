// Code generated by protoc-gen-go. DO NOT EDIT.
// source: coin.proto

package matpool

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

type CoinType int32

const (
	CoinType_InvalidCoin CoinType = 0
	CoinType_EY        CoinType = 1
)

var CoinType_name = map[int32]string{
	0: "InvalidCoin",
	1: "EY",
}

var CoinType_value = map[string]int32{
	"InvalidCoin": 0,
	"EY":        1,
}

func (x CoinType) String() string {
	return proto.EnumName(CoinType_name, int32(x))
}

func (CoinType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_39141bafd5884f78, []int{0}
}

func init() {
	proto.RegisterEnum("matpool.CoinType", CoinType_name, CoinType_value)
}

func init() { proto.RegisterFile("coin.proto", fileDescriptor_39141bafd5884f78) }

var fileDescriptor_39141bafd5884f78 = []byte{
	// 87 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4a, 0xce, 0xcf, 0xcc,
	0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xcf, 0x4d, 0x2c, 0x29, 0xc8, 0xcf, 0xcf, 0xd1,
	0x52, 0xe5, 0xe2, 0x70, 0xce, 0xcf, 0xcc, 0x0b, 0xa9, 0x2c, 0x48, 0x15, 0xe2, 0xe7, 0xe2, 0xf6,
	0xcc, 0x2b, 0x4b, 0xcc, 0xc9, 0x4c, 0x01, 0x09, 0x09, 0x30, 0x08, 0x71, 0x70, 0xb1, 0x38, 0x85,
	0xf8, 0x3a, 0x0b, 0x30, 0x26, 0xb1, 0x81, 0xb5, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x12,
	0x16, 0x4b, 0x2b, 0x44, 0x00, 0x00, 0x00,
}
