// Code generated by protoc-gen-go. DO NOT EDIT.
// source: chaincode_event.proto

package peer

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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

//ChaincodeEvent is used for events and registrations that are specific to chaincode
//string type - "chaincode"
type ChaincodeEvent struct {
	ChaincodeId          string   `protobuf:"bytes,1,opt,name=chaincode_id,json=chaincodeId,proto3" json:"chaincode_id,omitempty"`
	TxId                 string   `protobuf:"bytes,2,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	EventName            string   `protobuf:"bytes,3,opt,name=event_name,json=eventName,proto3" json:"event_name,omitempty"`
	Payload              []byte   `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChaincodeEvent) Reset()         { *m = ChaincodeEvent{} }
func (m *ChaincodeEvent) String() string { return proto.CompactTextString(m) }
func (*ChaincodeEvent) ProtoMessage()    {}
func (*ChaincodeEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_0aa6411063813fab, []int{0}
}

func (m *ChaincodeEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeEvent.Unmarshal(m, b)
}
func (m *ChaincodeEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeEvent.Marshal(b, m, deterministic)
}
func (m *ChaincodeEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeEvent.Merge(m, src)
}
func (m *ChaincodeEvent) XXX_Size() int {
	return xxx_messageInfo_ChaincodeEvent.Size(m)
}
func (m *ChaincodeEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeEvent.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeEvent proto.InternalMessageInfo

func (m *ChaincodeEvent) GetChaincodeId() string {
	if m != nil {
		return m.ChaincodeId
	}
	return ""
}

func (m *ChaincodeEvent) GetTxId() string {
	if m != nil {
		return m.TxId
	}
	return ""
}

func (m *ChaincodeEvent) GetEventName() string {
	if m != nil {
		return m.EventName
	}
	return ""
}

func (m *ChaincodeEvent) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func init() {
	proto.RegisterType((*ChaincodeEvent)(nil), "protos.ChaincodeEvent")
}

func init() { proto.RegisterFile("chaincode_event.proto", fileDescriptor_0aa6411063813fab) }

var fileDescriptor_0aa6411063813fab = []byte{
	// 219 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4d, 0xce, 0x48, 0xcc,
	0xcc, 0x4b, 0xce, 0x4f, 0x49, 0x8d, 0x4f, 0x2d, 0x4b, 0xcd, 0x2b, 0xd1, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x17, 0x62, 0x03, 0x53, 0xc5, 0x4a, 0x8d, 0x8c, 0x5c, 0x7c, 0xce, 0x30, 0x15, 0xae, 0x20,
	0x05, 0x42, 0x8a, 0x5c, 0x3c, 0x08, 0x3d, 0x99, 0x29, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41,
	0xdc, 0x70, 0x31, 0xcf, 0x14, 0x21, 0x61, 0x2e, 0xd6, 0x92, 0x0a, 0x90, 0x1c, 0x13, 0x58, 0x8e,
	0xa5, 0xa4, 0xc2, 0x33, 0x45, 0x48, 0x96, 0x8b, 0x0b, 0x6c, 0x43, 0x7c, 0x5e, 0x62, 0x6e, 0xaa,
	0x04, 0x33, 0x58, 0x86, 0x13, 0x2c, 0xe2, 0x97, 0x98, 0x9b, 0x2a, 0x24, 0xc1, 0xc5, 0x5e, 0x90,
	0x58, 0x99, 0x93, 0x9f, 0x98, 0x22, 0xc1, 0xa2, 0xc0, 0xa8, 0xc1, 0x13, 0x04, 0xe3, 0x3a, 0xa5,
	0x71, 0x71, 0x43, 0x5c, 0xa3, 0x57, 0x90, 0x9a, 0x5a, 0xe4, 0x24, 0x8a, 0xea, 0x9e, 0x80, 0xc4,
	0xe4, 0xec, 0xc4, 0xf4, 0xd4, 0x28, 0x87, 0xf4, 0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd, 0xe4, 0xfc,
	0x5c, 0xfd, 0x82, 0xc4, 0x9c, 0x9c, 0xd4, 0x92, 0xfc, 0xbc, 0x54, 0xfd, 0xf4, 0x7c, 0x5d, 0x04,
	0x27, 0x39, 0xbf, 0x28, 0x55, 0xbf, 0x2c, 0xd7, 0x39, 0x3f, 0xaf, 0xa4, 0x28, 0x31, 0xb9, 0x24,
	0xa0, 0x34, 0x49, 0x1f, 0x62, 0xb0, 0x3e, 0xc8, 0xe0, 0x24, 0x88, 0x9f, 0x8d, 0x01, 0x01, 0x00,
	0x00, 0xff, 0xff, 0xd1, 0x5a, 0x7f, 0x33, 0x13, 0x01, 0x00, 0x00,
}