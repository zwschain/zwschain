// Code generated by protoc-gen-go. DO NOT EDIT.
// source: chaincode.proto

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

// Confidentiality Levels
type ConfidentialityLevel int32

const (
	ConfidentialityLevel_PUBLIC       ConfidentialityLevel = 0
	ConfidentialityLevel_CONFIDENTIAL ConfidentialityLevel = 1
)

var ConfidentialityLevel_name = map[int32]string{
	0: "PUBLIC",
	1: "CONFIDENTIAL",
}

var ConfidentialityLevel_value = map[string]int32{
	"PUBLIC":       0,
	"CONFIDENTIAL": 1,
}

func (x ConfidentialityLevel) String() string {
	return proto.EnumName(ConfidentialityLevel_name, int32(x))
}

func (ConfidentialityLevel) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{0}
}

type ChaincodeSpec_Type int32

const (
	ChaincodeSpec_UNDEFINED ChaincodeSpec_Type = 0
	ChaincodeSpec_GOLANG    ChaincodeSpec_Type = 1
	ChaincodeSpec_NODE      ChaincodeSpec_Type = 2
	ChaincodeSpec_CAR       ChaincodeSpec_Type = 3
	ChaincodeSpec_JAVA      ChaincodeSpec_Type = 4
)

var ChaincodeSpec_Type_name = map[int32]string{
	0: "UNDEFINED",
	1: "GOLANG",
	2: "NODE",
	3: "CAR",
	4: "JAVA",
}

var ChaincodeSpec_Type_value = map[string]int32{
	"UNDEFINED": 0,
	"GOLANG":    1,
	"NODE":      2,
	"CAR":       3,
	"JAVA":      4,
}

func (x ChaincodeSpec_Type) String() string {
	return proto.EnumName(ChaincodeSpec_Type_name, int32(x))
}

func (ChaincodeSpec_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{2, 0}
}

type ChaincodeDeploymentSpec_ExecutionEnvironment int32

const (
	ChaincodeDeploymentSpec_DOCKER ChaincodeDeploymentSpec_ExecutionEnvironment = 0
	ChaincodeDeploymentSpec_SYSTEM ChaincodeDeploymentSpec_ExecutionEnvironment = 1
)

var ChaincodeDeploymentSpec_ExecutionEnvironment_name = map[int32]string{
	0: "DOCKER",
	1: "SYSTEM",
}

var ChaincodeDeploymentSpec_ExecutionEnvironment_value = map[string]int32{
	"DOCKER": 0,
	"SYSTEM": 1,
}

func (x ChaincodeDeploymentSpec_ExecutionEnvironment) String() string {
	return proto.EnumName(ChaincodeDeploymentSpec_ExecutionEnvironment_name, int32(x))
}

func (ChaincodeDeploymentSpec_ExecutionEnvironment) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{3, 0}
}

//ChaincodeID contains the path as specified by the deploy transaction
//that created it as well as the hashCode that is generated by the
//system for the path. From the user level (ie, CLI, REST API and so on)
//deploy transaction is expected to provide the path and other requests
//are expected to provide the hashCode. The other value will be ignored.
//Internally, the structure could contain both values. For instance, the
//hashCode will be set when first generated using the path
type ChaincodeID struct {
	//deploy transaction will use the path
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	//all other requests will use the name (really a hashcode) generated by
	//the deploy transaction
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	//user friendly version name for the chaincode
	Version              string   `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChaincodeID) Reset()         { *m = ChaincodeID{} }
func (m *ChaincodeID) String() string { return proto.CompactTextString(m) }
func (*ChaincodeID) ProtoMessage()    {}
func (*ChaincodeID) Descriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{0}
}

func (m *ChaincodeID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeID.Unmarshal(m, b)
}
func (m *ChaincodeID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeID.Marshal(b, m, deterministic)
}
func (m *ChaincodeID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeID.Merge(m, src)
}
func (m *ChaincodeID) XXX_Size() int {
	return xxx_messageInfo_ChaincodeID.Size(m)
}
func (m *ChaincodeID) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeID.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeID proto.InternalMessageInfo

func (m *ChaincodeID) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *ChaincodeID) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ChaincodeID) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

// Carries the chaincode function and its arguments.
// UnmarshalJSON in transaction.go converts the string-based REST/JSON input to
// the []byte-based current ChaincodeInput structure.
type ChaincodeInput struct {
	Args                 [][]byte          `protobuf:"bytes,1,rep,name=args,proto3" json:"args,omitempty"`
	Decorations          map[string][]byte `protobuf:"bytes,2,rep,name=decorations,proto3" json:"decorations,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ChaincodeInput) Reset()         { *m = ChaincodeInput{} }
func (m *ChaincodeInput) String() string { return proto.CompactTextString(m) }
func (*ChaincodeInput) ProtoMessage()    {}
func (*ChaincodeInput) Descriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{1}
}

func (m *ChaincodeInput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeInput.Unmarshal(m, b)
}
func (m *ChaincodeInput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeInput.Marshal(b, m, deterministic)
}
func (m *ChaincodeInput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeInput.Merge(m, src)
}
func (m *ChaincodeInput) XXX_Size() int {
	return xxx_messageInfo_ChaincodeInput.Size(m)
}
func (m *ChaincodeInput) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeInput.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeInput proto.InternalMessageInfo

func (m *ChaincodeInput) GetArgs() [][]byte {
	if m != nil {
		return m.Args
	}
	return nil
}

func (m *ChaincodeInput) GetDecorations() map[string][]byte {
	if m != nil {
		return m.Decorations
	}
	return nil
}

// Carries the chaincode specification. This is the actual metadata required for
// defining a chaincode.
type ChaincodeSpec struct {
	Type                 ChaincodeSpec_Type `protobuf:"varint,1,opt,name=type,proto3,enum=protos.ChaincodeSpec_Type" json:"type,omitempty"`
	ChaincodeId          *ChaincodeID       `protobuf:"bytes,2,opt,name=chaincode_id,json=chaincodeId,proto3" json:"chaincode_id,omitempty"`
	Input                *ChaincodeInput    `protobuf:"bytes,3,opt,name=input,proto3" json:"input,omitempty"`
	Timeout              int32              `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
	Memory               int64              `protobuf:"varint,5,opt,name=Memory,proto3" json:"Memory,omitempty"`
	CpuQuota             int64              `protobuf:"varint,6,opt,name=CpuQuota,proto3" json:"CpuQuota,omitempty"`
	CpuShare             int64              `protobuf:"varint,7,opt,name=CpuShare,proto3" json:"CpuShare,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *ChaincodeSpec) Reset()         { *m = ChaincodeSpec{} }
func (m *ChaincodeSpec) String() string { return proto.CompactTextString(m) }
func (*ChaincodeSpec) ProtoMessage()    {}
func (*ChaincodeSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{2}
}

func (m *ChaincodeSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeSpec.Unmarshal(m, b)
}
func (m *ChaincodeSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeSpec.Marshal(b, m, deterministic)
}
func (m *ChaincodeSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeSpec.Merge(m, src)
}
func (m *ChaincodeSpec) XXX_Size() int {
	return xxx_messageInfo_ChaincodeSpec.Size(m)
}
func (m *ChaincodeSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeSpec proto.InternalMessageInfo

func (m *ChaincodeSpec) GetType() ChaincodeSpec_Type {
	if m != nil {
		return m.Type
	}
	return ChaincodeSpec_UNDEFINED
}

func (m *ChaincodeSpec) GetChaincodeId() *ChaincodeID {
	if m != nil {
		return m.ChaincodeId
	}
	return nil
}

func (m *ChaincodeSpec) GetInput() *ChaincodeInput {
	if m != nil {
		return m.Input
	}
	return nil
}

func (m *ChaincodeSpec) GetTimeout() int32 {
	if m != nil {
		return m.Timeout
	}
	return 0
}

func (m *ChaincodeSpec) GetMemory() int64 {
	if m != nil {
		return m.Memory
	}
	return 0
}

func (m *ChaincodeSpec) GetCpuQuota() int64 {
	if m != nil {
		return m.CpuQuota
	}
	return 0
}

func (m *ChaincodeSpec) GetCpuShare() int64 {
	if m != nil {
		return m.CpuShare
	}
	return 0
}

// Specify the deployment of a chaincode.
// TODO: Define `codePackage`.
type ChaincodeDeploymentSpec struct {
	ChaincodeSpec *ChaincodeSpec `protobuf:"bytes,1,opt,name=chaincode_spec,json=chaincodeSpec,proto3" json:"chaincode_spec,omitempty"`
	// Controls when the chaincode becomes executable.
	//    google.protobuf.Timestamp effective_date = 2;
	CodePackage          []byte                                       `protobuf:"bytes,3,opt,name=code_package,json=codePackage,proto3" json:"code_package,omitempty"`
	ExecEnv              ChaincodeDeploymentSpec_ExecutionEnvironment `protobuf:"varint,4,opt,name=exec_env,json=execEnv,proto3,enum=protos.ChaincodeDeploymentSpec_ExecutionEnvironment" json:"exec_env,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                     `json:"-"`
	XXX_unrecognized     []byte                                       `json:"-"`
	XXX_sizecache        int32                                        `json:"-"`
}

func (m *ChaincodeDeploymentSpec) Reset()         { *m = ChaincodeDeploymentSpec{} }
func (m *ChaincodeDeploymentSpec) String() string { return proto.CompactTextString(m) }
func (*ChaincodeDeploymentSpec) ProtoMessage()    {}
func (*ChaincodeDeploymentSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{3}
}

func (m *ChaincodeDeploymentSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeDeploymentSpec.Unmarshal(m, b)
}
func (m *ChaincodeDeploymentSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeDeploymentSpec.Marshal(b, m, deterministic)
}
func (m *ChaincodeDeploymentSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeDeploymentSpec.Merge(m, src)
}
func (m *ChaincodeDeploymentSpec) XXX_Size() int {
	return xxx_messageInfo_ChaincodeDeploymentSpec.Size(m)
}
func (m *ChaincodeDeploymentSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeDeploymentSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeDeploymentSpec proto.InternalMessageInfo

func (m *ChaincodeDeploymentSpec) GetChaincodeSpec() *ChaincodeSpec {
	if m != nil {
		return m.ChaincodeSpec
	}
	return nil
}

func (m *ChaincodeDeploymentSpec) GetCodePackage() []byte {
	if m != nil {
		return m.CodePackage
	}
	return nil
}

func (m *ChaincodeDeploymentSpec) GetExecEnv() ChaincodeDeploymentSpec_ExecutionEnvironment {
	if m != nil {
		return m.ExecEnv
	}
	return ChaincodeDeploymentSpec_DOCKER
}

// Carries the chaincode function and its arguments.
type ChaincodeInvocationSpec struct {
	ChaincodeSpec *ChaincodeSpec `protobuf:"bytes,1,opt,name=chaincode_spec,json=chaincodeSpec,proto3" json:"chaincode_spec,omitempty"`
	// This field can contain a user-specified ID generation algorithm
	// If supplied, this will be used to generate a ID
	// If not supplied (left empty), sha256base64 will be used
	// The algorithm consists of two parts:
	//  1, a hash function
	//  2, a decoding used to decode user (string) input to bytes
	// Currently, SHA256 with BASE64 is supported (e.g. idGenerationAlg='sha256base64')
	IdGenerationAlg      string   `protobuf:"bytes,2,opt,name=id_generation_alg,json=idGenerationAlg,proto3" json:"id_generation_alg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChaincodeInvocationSpec) Reset()         { *m = ChaincodeInvocationSpec{} }
func (m *ChaincodeInvocationSpec) String() string { return proto.CompactTextString(m) }
func (*ChaincodeInvocationSpec) ProtoMessage()    {}
func (*ChaincodeInvocationSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_97136ef4b384cc22, []int{4}
}

func (m *ChaincodeInvocationSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChaincodeInvocationSpec.Unmarshal(m, b)
}
func (m *ChaincodeInvocationSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChaincodeInvocationSpec.Marshal(b, m, deterministic)
}
func (m *ChaincodeInvocationSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChaincodeInvocationSpec.Merge(m, src)
}
func (m *ChaincodeInvocationSpec) XXX_Size() int {
	return xxx_messageInfo_ChaincodeInvocationSpec.Size(m)
}
func (m *ChaincodeInvocationSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ChaincodeInvocationSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ChaincodeInvocationSpec proto.InternalMessageInfo

func (m *ChaincodeInvocationSpec) GetChaincodeSpec() *ChaincodeSpec {
	if m != nil {
		return m.ChaincodeSpec
	}
	return nil
}

func (m *ChaincodeInvocationSpec) GetIdGenerationAlg() string {
	if m != nil {
		return m.IdGenerationAlg
	}
	return ""
}

func init() {
	proto.RegisterEnum("protos.ConfidentialityLevel", ConfidentialityLevel_name, ConfidentialityLevel_value)
	proto.RegisterEnum("protos.ChaincodeSpec_Type", ChaincodeSpec_Type_name, ChaincodeSpec_Type_value)
	proto.RegisterEnum("protos.ChaincodeDeploymentSpec_ExecutionEnvironment", ChaincodeDeploymentSpec_ExecutionEnvironment_name, ChaincodeDeploymentSpec_ExecutionEnvironment_value)
	proto.RegisterType((*ChaincodeID)(nil), "protos.ChaincodeID")
	proto.RegisterType((*ChaincodeInput)(nil), "protos.ChaincodeInput")
	proto.RegisterMapType((map[string][]byte)(nil), "protos.ChaincodeInput.DecorationsEntry")
	proto.RegisterType((*ChaincodeSpec)(nil), "protos.ChaincodeSpec")
	proto.RegisterType((*ChaincodeDeploymentSpec)(nil), "protos.ChaincodeDeploymentSpec")
	proto.RegisterType((*ChaincodeInvocationSpec)(nil), "protos.ChaincodeInvocationSpec")
}

func init() { proto.RegisterFile("chaincode.proto", fileDescriptor_97136ef4b384cc22) }

var fileDescriptor_97136ef4b384cc22 = []byte{
	// 641 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xad, 0xf3, 0xd9, 0x8e, 0xd3, 0xd4, 0x2c, 0xa5, 0x58, 0x3d, 0x85, 0x5c, 0x88, 0x2a, 0x70,
	0xa4, 0x50, 0x21, 0x84, 0x50, 0x45, 0x6a, 0xbb, 0x95, 0x21, 0x4d, 0x8a, 0xd3, 0x22, 0xc1, 0x25,
	0xda, 0x3a, 0x43, 0x62, 0xd5, 0xd9, 0xb5, 0x9c, 0xb5, 0xd5, 0x9c, 0xf9, 0x41, 0xfc, 0x14, 0x7e,
	0x10, 0x17, 0xb4, 0xeb, 0x34, 0x69, 0x69, 0x8f, 0x9c, 0x32, 0xf3, 0xe6, 0xed, 0xdb, 0x79, 0x2f,
	0x9b, 0xc0, 0x4e, 0x30, 0xa5, 0x21, 0x0b, 0xf8, 0x18, 0xad, 0x38, 0xe1, 0x82, 0x93, 0x8a, 0xfa,
	0x98, 0x37, 0x07, 0xa0, 0xdb, 0xb7, 0x23, 0xcf, 0x21, 0x04, 0x4a, 0x31, 0x15, 0x53, 0x53, 0x6b,
	0x68, 0xad, 0x2d, 0x5f, 0xd5, 0x12, 0x63, 0x74, 0x86, 0x66, 0x21, 0xc7, 0x64, 0x4d, 0x4c, 0xa8,
	0x66, 0x98, 0xcc, 0x43, 0xce, 0xcc, 0xa2, 0x82, 0x6f, 0xdb, 0xe6, 0x2f, 0x0d, 0xea, 0x6b, 0x45,
	0x16, 0xa7, 0x42, 0x0a, 0xd0, 0x64, 0x32, 0x37, 0xb5, 0x46, 0xb1, 0x55, 0xf3, 0x55, 0x4d, 0x3c,
	0xd0, 0xc7, 0x18, 0xf0, 0x84, 0x8a, 0x90, 0xb3, 0xb9, 0x59, 0x68, 0x14, 0x5b, 0x7a, 0xe7, 0x65,
	0xbe, 0xdc, 0xdc, 0xba, 0x2f, 0x60, 0x39, 0x6b, 0xa6, 0xcb, 0x44, 0xb2, 0xf0, 0xef, 0x9e, 0xdd,
	0x3f, 0x02, 0xe3, 0x5f, 0x02, 0x31, 0xa0, 0x78, 0x8d, 0x8b, 0xa5, 0x0d, 0x59, 0x92, 0x5d, 0x28,
	0x67, 0x34, 0x4a, 0x73, 0x1b, 0x35, 0x3f, 0x6f, 0xde, 0x17, 0xde, 0x69, 0xcd, 0xdf, 0x05, 0xd8,
	0x5e, 0x5d, 0x38, 0x8c, 0x31, 0x20, 0x16, 0x94, 0xc4, 0x22, 0x46, 0x75, 0xbc, 0xde, 0xd9, 0x7f,
	0xb0, 0x95, 0x24, 0x59, 0x17, 0x8b, 0x18, 0x7d, 0xc5, 0x23, 0x6f, 0xa1, 0xb6, 0xca, 0x77, 0x14,
	0x8e, 0xd5, 0x15, 0x7a, 0xe7, 0xe9, 0x43, 0x37, 0x8e, 0xaf, 0xaf, 0x88, 0xde, 0x98, 0xbc, 0x82,
	0x72, 0x28, 0x0d, 0xaa, 0x0c, 0xf5, 0xce, 0xde, 0xe3, 0xf6, 0xfd, 0x9c, 0x24, 0x33, 0x17, 0xe1,
	0x0c, 0x79, 0x2a, 0xcc, 0x52, 0x43, 0x6b, 0x95, 0xfd, 0xdb, 0x96, 0xec, 0x41, 0xe5, 0x0c, 0x67,
	0x3c, 0x59, 0x98, 0xe5, 0x86, 0xd6, 0x2a, 0xfa, 0xcb, 0x8e, 0xec, 0xc3, 0xa6, 0x1d, 0xa7, 0x5f,
	0x52, 0x2e, 0xa8, 0x59, 0x51, 0x93, 0x55, 0xbf, 0x9c, 0x0d, 0xa7, 0x34, 0x41, 0xb3, 0xba, 0x9a,
	0xa9, 0xbe, 0x79, 0x04, 0x25, 0xe9, 0x8e, 0x6c, 0xc3, 0xd6, 0x65, 0xdf, 0x71, 0x4f, 0xbc, 0xbe,
	0xeb, 0x18, 0x1b, 0x04, 0xa0, 0x72, 0x3a, 0xe8, 0x75, 0xfb, 0xa7, 0x86, 0x46, 0x36, 0xa1, 0xd4,
	0x1f, 0x38, 0xae, 0x51, 0x20, 0x55, 0x28, 0xda, 0x5d, 0xdf, 0x28, 0x4a, 0xe8, 0x53, 0xf7, 0x6b,
	0xd7, 0x28, 0x35, 0xff, 0x68, 0xf0, 0x7c, 0xe5, 0xc1, 0xc1, 0x38, 0xe2, 0x8b, 0x19, 0x32, 0xa1,
	0xb2, 0xfd, 0x00, 0xf5, 0x75, 0x56, 0xf3, 0x18, 0x03, 0x95, 0xb2, 0xde, 0x79, 0xf6, 0x68, 0xca,
	0xfe, 0x76, 0x70, 0xef, 0x9b, 0x79, 0x01, 0x35, 0x75, 0x30, 0xa6, 0xc1, 0x35, 0x9d, 0xa0, 0x0a,
	0xae, 0xe6, 0xeb, 0x12, 0x3b, 0xcf, 0x21, 0x32, 0x80, 0x4d, 0xbc, 0xc1, 0x60, 0x84, 0x2c, 0x53,
	0x39, 0xd5, 0x3b, 0x87, 0x0f, 0xa4, 0xef, 0xef, 0x64, 0xb9, 0x37, 0x18, 0xa4, 0xf2, 0xf5, 0xb8,
	0x2c, 0x0b, 0x13, 0xce, 0xe4, 0xc0, 0xaf, 0x4a, 0x15, 0x97, 0x65, 0x4d, 0x0b, 0x76, 0x1f, 0x23,
	0xc8, 0x38, 0x9c, 0x81, 0xfd, 0xd9, 0xf5, 0xf3, 0x68, 0x86, 0xdf, 0x86, 0x17, 0xee, 0x99, 0xa1,
	0x35, 0x7f, 0xde, 0x75, 0xef, 0xb1, 0x8c, 0x07, 0xea, 0x65, 0xfe, 0x07, 0xf7, 0x07, 0xf0, 0x24,
	0x1c, 0x8f, 0x26, 0xc8, 0x30, 0x7f, 0xec, 0x23, 0x1a, 0x4d, 0x96, 0x3f, 0xcb, 0x9d, 0x70, 0x7c,
	0xba, 0xc2, 0xbb, 0xd1, 0xe4, 0xe0, 0x10, 0x76, 0x6d, 0xce, 0x7e, 0x84, 0x63, 0x64, 0x22, 0xa4,
	0x51, 0x28, 0x16, 0x3d, 0xcc, 0x30, 0x92, 0x9b, 0x9e, 0x5f, 0x1e, 0xf7, 0x3c, 0xdb, 0xd8, 0x20,
	0x06, 0xd4, 0xec, 0x41, 0xff, 0xc4, 0x73, 0xdc, 0xfe, 0x85, 0xd7, 0xed, 0x19, 0xda, 0xf1, 0x00,
	0xf4, 0xe5, 0x22, 0x31, 0x62, 0xf2, 0xfd, 0xe3, 0x24, 0x14, 0xd3, 0xf4, 0xca, 0x0a, 0xf8, 0xac,
	0x1d, 0xd3, 0x28, 0x42, 0xc1, 0x19, 0xb6, 0x27, 0xfc, 0xf5, 0xba, 0x09, 0x78, 0x82, 0xed, 0x6c,
	0x66, 0x73, 0x26, 0x12, 0x1a, 0x88, 0xf3, 0xf4, 0xaa, 0x9d, 0x2b, 0xb4, 0xa5, 0xc2, 0x55, 0xfe,
	0x3f, 0xf3, 0xe6, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc2, 0x3a, 0xc3, 0x07, 0x81, 0x04, 0x00,
	0x00,
}
