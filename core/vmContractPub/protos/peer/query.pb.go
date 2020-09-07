// Code generated by protoc-gen-go. DO NOT EDIT.
// source: query.proto

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

// ChaincodeQueryResponse returns information about each chaincode that pertains
// to a query in lscc.go, such as GetChaincodes (returns all chaincodes
// instantiated on a channel), and GetInstalledChaincodes (returns all chaincodes
// installed on a peer)
type ChaincodeQueryResponse struct {
	Chaincodes           []*ChaincodeInfo `protobuf:"bytes,1,rep,name=chaincodes,proto3" json:"chaincodes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *ChaincodeQueryResponse) Reset()         { *m = ChaincodeQueryResponse{} }
func (m *ChaincodeQueryResponse) String() string { return proto.CompactTextString(m) }
func (*ChaincodeQueryResponse) ProtoMessage()    {}
func (*ChaincodeQueryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{0}
}

func (m *ChaincodeQueryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeQueryResponse.Unmarshal(m, b)
}
func (m *ChaincodeQueryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeQueryResponse.Marshal(b, m, deterministic)
}
func (m *ChaincodeQueryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeQueryResponse.Merge(m, src)
}
func (m *ChaincodeQueryResponse) XXX_Size() int {
	return xxx_messageInfo_ChaincodeQueryResponse.Size(m)
}
func (m *ChaincodeQueryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeQueryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeQueryResponse proto.InternalMessageInfo

func (m *ChaincodeQueryResponse) GetChaincodes() []*ChaincodeInfo {
	if m != nil {
		return m.Chaincodes
	}
	return nil
}

// ChaincodeInfo contains general information about an installed/instantiated
// chaincode
type ChaincodeInfo struct {
	Name    string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// the path as specified by the install/instantiate transaction
	Path string `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	// the chaincode function upon instantiation and its arguments. This will be
	// blank if the query is returning information about installed chaincodes.
	Input string `protobuf:"bytes,4,opt,name=input,proto3" json:"input,omitempty"`
	// the name of the ESCC for this chaincode. This will be
	// blank if the query is returning information about installed chaincodes.
	Escc string `protobuf:"bytes,5,opt,name=escc,proto3" json:"escc,omitempty"`
	// the name of the VSCC for this chaincode. This will be
	// blank if the query is returning information about installed chaincodes.
	Vscc string `protobuf:"bytes,6,opt,name=vscc,proto3" json:"vscc,omitempty"`
	// the chaincode unique id.
	// computed as: H(
	//                H(name || version) ||
	//                H(CodePackage)
	//              )
	Id                   []byte   `protobuf:"bytes,7,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChaincodeInfo) Reset()         { *m = ChaincodeInfo{} }
func (m *ChaincodeInfo) String() string { return proto.CompactTextString(m) }
func (*ChaincodeInfo) ProtoMessage()    {}
func (*ChaincodeInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{1}
}

func (m *ChaincodeInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeInfo.Unmarshal(m, b)
}
func (m *ChaincodeInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeInfo.Marshal(b, m, deterministic)
}
func (m *ChaincodeInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeInfo.Merge(m, src)
}
func (m *ChaincodeInfo) XXX_Size() int {
	return xxx_messageInfo_ChaincodeInfo.Size(m)
}
func (m *ChaincodeInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeInfo proto.InternalMessageInfo

func (m *ChaincodeInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ChaincodeInfo) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *ChaincodeInfo) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *ChaincodeInfo) GetInput() string {
	if m != nil {
		return m.Input
	}
	return ""
}

func (m *ChaincodeInfo) GetEscc() string {
	if m != nil {
		return m.Escc
	}
	return ""
}

func (m *ChaincodeInfo) GetVscc() string {
	if m != nil {
		return m.Vscc
	}
	return ""
}

func (m *ChaincodeInfo) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

// ChannelQueryResponse returns information about each channel that pertains
// to a query in lscc.go, such as GetChannels (returns all channels for a
// given peer)
type ChannelQueryResponse struct {
	Channels             []*ChannelInfo `protobuf:"bytes,1,rep,name=channels,proto3" json:"channels,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ChannelQueryResponse) Reset()         { *m = ChannelQueryResponse{} }
func (m *ChannelQueryResponse) String() string { return proto.CompactTextString(m) }
func (*ChannelQueryResponse) ProtoMessage()    {}
func (*ChannelQueryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{2}
}

func (m *ChannelQueryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChannelQueryResponse.Unmarshal(m, b)
}
func (m *ChannelQueryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChannelQueryResponse.Marshal(b, m, deterministic)
}
func (m *ChannelQueryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChannelQueryResponse.Merge(m, src)
}
func (m *ChannelQueryResponse) XXX_Size() int {
	return xxx_messageInfo_ChannelQueryResponse.Size(m)
}
func (m *ChannelQueryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ChannelQueryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ChannelQueryResponse proto.InternalMessageInfo

func (m *ChannelQueryResponse) GetChannels() []*ChannelInfo {
	if m != nil {
		return m.Channels
	}
	return nil
}

// ChannelInfo contains general information about channels
type ChannelInfo struct {
	ChannelId            string   `protobuf:"bytes,1,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChannelInfo) Reset()         { *m = ChannelInfo{} }
func (m *ChannelInfo) String() string { return proto.CompactTextString(m) }
func (*ChannelInfo) ProtoMessage()    {}
func (*ChannelInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{3}
}

func (m *ChannelInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChannelInfo.Unmarshal(m, b)
}
func (m *ChannelInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChannelInfo.Marshal(b, m, deterministic)
}
func (m *ChannelInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChannelInfo.Merge(m, src)
}
func (m *ChannelInfo) XXX_Size() int {
	return xxx_messageInfo_ChannelInfo.Size(m)
}
func (m *ChannelInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ChannelInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ChannelInfo proto.InternalMessageInfo

func (m *ChannelInfo) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func init() {
	proto.RegisterType((*ChaincodeQueryResponse)(nil), "protos.ChaincodeQueryResponse")
	proto.RegisterType((*ChaincodeInfo)(nil), "protos.ChaincodeInfo")
	proto.RegisterType((*ChannelQueryResponse)(nil), "protos.ChannelQueryResponse")
	proto.RegisterType((*ChannelInfo)(nil), "protos.ChannelInfo")
}

func init() { proto.RegisterFile("query.proto", fileDescriptor_5c6ac9b241082464) }

var fileDescriptor_5c6ac9b241082464 = []byte{
	// 298 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x91, 0x3f, 0x4f, 0xc3, 0x30,
	0x10, 0xc5, 0x95, 0xfe, 0xa5, 0x17, 0x60, 0x30, 0x05, 0x79, 0x41, 0xaa, 0x32, 0x75, 0x80, 0x5a,
	0x02, 0xb1, 0x23, 0x3a, 0xa0, 0x4e, 0x85, 0x8c, 0x2c, 0xc8, 0x75, 0x8e, 0xc6, 0x52, 0x6b, 0x1b,
	0xdb, 0xa9, 0xc4, 0xa7, 0xe1, 0xab, 0x22, 0xdb, 0x6d, 0x49, 0xa7, 0xdc, 0xfb, 0xbd, 0x77, 0x27,
	0xbd, 0x18, 0xf2, 0xef, 0x06, 0xed, 0xcf, 0xcc, 0x58, 0xed, 0x35, 0x19, 0xc4, 0x8f, 0x2b, 0x96,
	0x70, 0x33, 0xaf, 0xb9, 0x54, 0x42, 0x57, 0xf8, 0x1e, 0xfc, 0x12, 0x9d, 0xd1, 0xca, 0x21, 0x79,
	0x02, 0x10, 0x07, 0xc7, 0xd1, 0x6c, 0xd2, 0x9d, 0xe6, 0x0f, 0xd7, 0x69, 0xdb, 0xcd, 0x8e, 0x3b,
	0x0b, 0xf5, 0xa5, 0xcb, 0x56, 0xb0, 0xf8, 0xcd, 0xe0, 0xe2, 0xc4, 0x25, 0x04, 0x7a, 0x8a, 0x6f,
	0x91, 0x66, 0x93, 0x6c, 0x3a, 0x2a, 0xe3, 0x4c, 0x28, 0x0c, 0x77, 0x68, 0x9d, 0xd4, 0x8a, 0x76,
	0x22, 0x3e, 0xc8, 0x90, 0x36, 0xdc, 0xd7, 0xb4, 0x9b, 0xd2, 0x61, 0x26, 0x63, 0xe8, 0x4b, 0x65,
	0x1a, 0x4f, 0x7b, 0x11, 0x26, 0x11, 0x92, 0xe8, 0x84, 0xa0, 0xfd, 0x94, 0x0c, 0x73, 0x60, 0xbb,
	0xc0, 0x06, 0x89, 0x85, 0x99, 0x5c, 0x42, 0x47, 0x56, 0x74, 0x38, 0xc9, 0xa6, 0xe7, 0x65, 0x47,
	0x56, 0xc5, 0x2b, 0x8c, 0xe7, 0x35, 0x57, 0x0a, 0x37, 0xa7, 0x85, 0x19, 0x9c, 0x89, 0xc4, 0x0f,
	0x75, 0xaf, 0x5a, 0x75, 0x03, 0x8f, 0x65, 0x8f, 0xa1, 0xe2, 0x0e, 0xf2, 0x96, 0x41, 0x6e, 0xe3,
	0x0f, 0x0b, 0xf2, 0x53, 0x56, 0xfb, 0xb6, 0xa3, 0x3d, 0x59, 0x54, 0x2f, 0x4b, 0xc8, 0xf7, 0xd7,
	0x0c, 0xa2, 0xfd, 0x78, 0x5e, 0x4b, 0x5f, 0x37, 0xab, 0x99, 0xd0, 0x5b, 0x66, 0xf8, 0x66, 0x83,
	0x5e, 0x2b, 0x64, 0x6b, 0x7d, 0xff, 0x2f, 0x84, 0xb6, 0xc8, 0x76, 0xdb, 0xb9, 0x56, 0xde, 0x72,
	0xe1, 0xdf, 0x9a, 0x15, 0x4b, 0x17, 0x58, 0xb8, 0xb0, 0x4a, 0x4f, 0xf8, 0xf8, 0x17, 0x00, 0x00,
	0xff, 0xff, 0x4d, 0x57, 0xa8, 0x15, 0xd8, 0x01, 0x00, 0x00,
}