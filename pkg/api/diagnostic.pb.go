// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: api/diagnostic.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Do a ping for a given hostname
type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host    string               `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Count   int32                `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	Timeout *durationpb.Duration `protobuf:"bytes,3,opt,name=timeout,proto3,oneof" json:"timeout,omitempty"`
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_diagnostic_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_diagnostic_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_api_diagnostic_proto_rawDescGZIP(), []int{0}
}

func (x *PingRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *PingRequest) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *PingRequest) GetTimeout() *durationpb.Duration {
	if x != nil {
		return x.Timeout
	}
	return nil
}

// Results of the ping test
type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pingable             bool                 `protobuf:"varint,1,opt,name=pingable,proto3" json:"pingable,omitempty"`
	PacketsReceived      int32                `protobuf:"varint,2,opt,name=packetsReceived,proto3" json:"packetsReceived,omitempty"`
	PacketsSent          int32                `protobuf:"varint,3,opt,name=packetsSent,proto3" json:"packetsSent,omitempty"`
	PacketLossPercentage int32                `protobuf:"varint,4,opt,name=packetLossPercentage,proto3" json:"packetLossPercentage,omitempty"`
	PingedHost           string               `protobuf:"bytes,5,opt,name=pingedHost,proto3" json:"pingedHost,omitempty"`
	MinRtt               *durationpb.Duration `protobuf:"bytes,6,opt,name=minRtt,proto3" json:"minRtt,omitempty"`
	MaxRtt               *durationpb.Duration `protobuf:"bytes,7,opt,name=maxRtt,proto3" json:"maxRtt,omitempty"`
	AvgRtt               *durationpb.Duration `protobuf:"bytes,8,opt,name=avgRtt,proto3" json:"avgRtt,omitempty"`
	StdDevRtt            *durationpb.Duration `protobuf:"bytes,9,opt,name=stdDevRtt,proto3" json:"stdDevRtt,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_diagnostic_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_diagnostic_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_api_diagnostic_proto_rawDescGZIP(), []int{1}
}

func (x *PingResponse) GetPingable() bool {
	if x != nil {
		return x.Pingable
	}
	return false
}

func (x *PingResponse) GetPacketsReceived() int32 {
	if x != nil {
		return x.PacketsReceived
	}
	return 0
}

func (x *PingResponse) GetPacketsSent() int32 {
	if x != nil {
		return x.PacketsSent
	}
	return 0
}

func (x *PingResponse) GetPacketLossPercentage() int32 {
	if x != nil {
		return x.PacketLossPercentage
	}
	return 0
}

func (x *PingResponse) GetPingedHost() string {
	if x != nil {
		return x.PingedHost
	}
	return ""
}

func (x *PingResponse) GetMinRtt() *durationpb.Duration {
	if x != nil {
		return x.MinRtt
	}
	return nil
}

func (x *PingResponse) GetMaxRtt() *durationpb.Duration {
	if x != nil {
		return x.MaxRtt
	}
	return nil
}

func (x *PingResponse) GetAvgRtt() *durationpb.Duration {
	if x != nil {
		return x.AvgRtt
	}
	return nil
}

func (x *PingResponse) GetStdDevRtt() *durationpb.Duration {
	if x != nil {
		return x.StdDevRtt
	}
	return nil
}

type PortProbeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Hostname or IP
	Host    string               `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Port    int32                `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Timeout *durationpb.Duration `protobuf:"bytes,3,opt,name=timeout,proto3,oneof" json:"timeout,omitempty"`
}

func (x *PortProbeRequest) Reset() {
	*x = PortProbeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_diagnostic_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortProbeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortProbeRequest) ProtoMessage() {}

func (x *PortProbeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_diagnostic_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortProbeRequest.ProtoReflect.Descriptor instead.
func (*PortProbeRequest) Descriptor() ([]byte, []int) {
	return file_api_diagnostic_proto_rawDescGZIP(), []int{2}
}

func (x *PortProbeRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *PortProbeRequest) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *PortProbeRequest) GetTimeout() *durationpb.Duration {
	if x != nil {
		return x.Timeout
	}
	return nil
}

type PortProbeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Open       bool    `protobuf:"varint,1,opt,name=open,proto3" json:"open,omitempty"`
	AddrProbed *string `protobuf:"bytes,2,opt,name=addrProbed,proto3,oneof" json:"addrProbed,omitempty"`
}

func (x *PortProbeResponse) Reset() {
	*x = PortProbeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_diagnostic_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortProbeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortProbeResponse) ProtoMessage() {}

func (x *PortProbeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_diagnostic_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortProbeResponse.ProtoReflect.Descriptor instead.
func (*PortProbeResponse) Descriptor() ([]byte, []int) {
	return file_api_diagnostic_proto_rawDescGZIP(), []int{3}
}

func (x *PortProbeResponse) GetOpen() bool {
	if x != nil {
		return x.Open
	}
	return false
}

func (x *PortProbeResponse) GetAddrProbed() string {
	if x != nil && x.AddrProbed != nil {
		return *x.AddrProbed
	}
	return ""
}

var File_api_diagnostic_proto protoreflect.FileDescriptor

var file_api_diagnostic_proto_rawDesc = []byte{
	0x0a, 0x14, 0x61, 0x70, 0x69, 0x2f, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61, 0x70, 0x69, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7d, 0x0a, 0x0b, 0x50,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f,
	0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x48, 0x00, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x88, 0x01, 0x01, 0x42, 0x0a,
	0x0a, 0x08, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x22, 0x9c, 0x03, 0x0a, 0x0c, 0x50,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x69, 0x6e, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x70,
	0x69, 0x6e, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x61, 0x63, 0x6b, 0x65,
	0x74, 0x73, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0f, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
	0x64, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x53, 0x65, 0x6e, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x73, 0x53,
	0x65, 0x6e, 0x74, 0x12, 0x32, 0x0a, 0x14, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x4c, 0x6f, 0x73,
	0x73, 0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x14, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x4c, 0x6f, 0x73, 0x73, 0x50, 0x65, 0x72,
	0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x69, 0x6e, 0x67, 0x65,
	0x64, 0x48, 0x6f, 0x73, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x69, 0x6e,
	0x67, 0x65, 0x64, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x31, 0x0a, 0x06, 0x6d, 0x69, 0x6e, 0x52, 0x74,
	0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x06, 0x6d, 0x69, 0x6e, 0x52, 0x74, 0x74, 0x12, 0x31, 0x0a, 0x06, 0x6d, 0x61,
	0x78, 0x52, 0x74, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x52, 0x74, 0x74, 0x12, 0x31, 0x0a,
	0x06, 0x61, 0x76, 0x67, 0x52, 0x74, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x06, 0x61, 0x76, 0x67, 0x52, 0x74, 0x74,
	0x12, 0x37, 0x0a, 0x09, 0x73, 0x74, 0x64, 0x44, 0x65, 0x76, 0x52, 0x74, 0x74, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09,
	0x73, 0x74, 0x64, 0x44, 0x65, 0x76, 0x52, 0x74, 0x74, 0x22, 0x80, 0x01, 0x0a, 0x10, 0x50, 0x6f,
	0x72, 0x74, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f,
	0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x38, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x48, 0x00, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x88, 0x01, 0x01,
	0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x22, 0x5b, 0x0a, 0x11,
	0x50, 0x6f, 0x72, 0x74, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x04, 0x6f, 0x70, 0x65, 0x6e, 0x12, 0x23, 0x0a, 0x0a, 0x61, 0x64, 0x64, 0x72, 0x50, 0x72, 0x6f,
	0x62, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0a, 0x61, 0x64, 0x64,
	0x72, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x64, 0x88, 0x01, 0x01, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x61,
	0x64, 0x64, 0x72, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x64, 0x32, 0x80, 0x01, 0x0a, 0x11, 0x44, 0x69,
	0x61, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x2d, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x10, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3c,
	0x0a, 0x09, 0x50, 0x6f, 0x72, 0x74, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x12, 0x15, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x50, 0x72, 0x6f,
	0x62, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x71, 0x0a, 0x07,
	0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x70, 0x69, 0x42, 0x0f, 0x44, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73,
	0x74, 0x69, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x61, 0x72, 0x74, 0x76, 0x65, 0x72, 0x6b, 0x65,
	0x74, 0x2f, 0x73, 0x6b, 0x69, 0x70, 0x63, 0x74, 0x6c, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x61, 0x70, 0x69, 0xa2, 0x02, 0x03, 0x41, 0x58, 0x58, 0xaa, 0x02, 0x03, 0x41, 0x70,
	0x69, 0xca, 0x02, 0x03, 0x41, 0x70, 0x69, 0xe2, 0x02, 0x0f, 0x41, 0x70, 0x69, 0x5c, 0x47, 0x50,
	0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x03, 0x41, 0x70, 0x69, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_diagnostic_proto_rawDescOnce sync.Once
	file_api_diagnostic_proto_rawDescData = file_api_diagnostic_proto_rawDesc
)

func file_api_diagnostic_proto_rawDescGZIP() []byte {
	file_api_diagnostic_proto_rawDescOnce.Do(func() {
		file_api_diagnostic_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_diagnostic_proto_rawDescData)
	})
	return file_api_diagnostic_proto_rawDescData
}

var file_api_diagnostic_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_api_diagnostic_proto_goTypes = []any{
	(*PingRequest)(nil),         // 0: api.PingRequest
	(*PingResponse)(nil),        // 1: api.PingResponse
	(*PortProbeRequest)(nil),    // 2: api.PortProbeRequest
	(*PortProbeResponse)(nil),   // 3: api.PortProbeResponse
	(*durationpb.Duration)(nil), // 4: google.protobuf.Duration
}
var file_api_diagnostic_proto_depIdxs = []int32{
	4, // 0: api.PingRequest.timeout:type_name -> google.protobuf.Duration
	4, // 1: api.PingResponse.minRtt:type_name -> google.protobuf.Duration
	4, // 2: api.PingResponse.maxRtt:type_name -> google.protobuf.Duration
	4, // 3: api.PingResponse.avgRtt:type_name -> google.protobuf.Duration
	4, // 4: api.PingResponse.stdDevRtt:type_name -> google.protobuf.Duration
	4, // 5: api.PortProbeRequest.timeout:type_name -> google.protobuf.Duration
	0, // 6: api.DiagnosticService.Ping:input_type -> api.PingRequest
	2, // 7: api.DiagnosticService.PortProbe:input_type -> api.PortProbeRequest
	1, // 8: api.DiagnosticService.Ping:output_type -> api.PingResponse
	3, // 9: api.DiagnosticService.PortProbe:output_type -> api.PortProbeResponse
	8, // [8:10] is the sub-list for method output_type
	6, // [6:8] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_api_diagnostic_proto_init() }
func file_api_diagnostic_proto_init() {
	if File_api_diagnostic_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_diagnostic_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*PingRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_diagnostic_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*PingResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_diagnostic_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*PortProbeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_diagnostic_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*PortProbeResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_api_diagnostic_proto_msgTypes[0].OneofWrappers = []any{}
	file_api_diagnostic_proto_msgTypes[2].OneofWrappers = []any{}
	file_api_diagnostic_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_diagnostic_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_diagnostic_proto_goTypes,
		DependencyIndexes: file_api_diagnostic_proto_depIdxs,
		MessageInfos:      file_api_diagnostic_proto_msgTypes,
	}.Build()
	File_api_diagnostic_proto = out.File
	file_api_diagnostic_proto_rawDesc = nil
	file_api_diagnostic_proto_goTypes = nil
	file_api_diagnostic_proto_depIdxs = nil
}