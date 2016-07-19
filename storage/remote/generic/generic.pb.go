// Code generated by protoc-gen-go.
// source: generic.proto
// DO NOT EDIT!

/*
Package generic is a generated protocol buffer package.

It is generated from these files:
	generic.proto

It has these top-level messages:
	Sample
	LabelPair
	TimeSeries
	GenericWriteRequest
	LabelMatcher
	GenericReadRequest
	GenericReadResponse
*/
package generic

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type MatchType int32

const (
	MatchType_EQUAL          MatchType = 0
	MatchType_NOT_EQUAL      MatchType = 1
	MatchType_REGEX_MATCH    MatchType = 2
	MatchType_REGEX_NO_MATCH MatchType = 3
)

var MatchType_name = map[int32]string{
	0: "EQUAL",
	1: "NOT_EQUAL",
	2: "REGEX_MATCH",
	3: "REGEX_NO_MATCH",
}
var MatchType_value = map[string]int32{
	"EQUAL":          0,
	"NOT_EQUAL":      1,
	"REGEX_MATCH":    2,
	"REGEX_NO_MATCH": 3,
}

func (x MatchType) Enum() *MatchType {
	p := new(MatchType)
	*p = x
	return p
}
func (x MatchType) String() string {
	return proto.EnumName(MatchType_name, int32(x))
}
func (x *MatchType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(MatchType_value, data, "MatchType")
	if err != nil {
		return err
	}
	*x = MatchType(value)
	return nil
}
func (MatchType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Sample struct {
	Value            *float64 `protobuf:"fixed64,1,opt,name=value" json:"value,omitempty"`
	TimestampMs      *int64   `protobuf:"varint,2,opt,name=timestamp_ms" json:"timestamp_ms,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Sample) Reset()                    { *m = Sample{} }
func (m *Sample) String() string            { return proto.CompactTextString(m) }
func (*Sample) ProtoMessage()               {}
func (*Sample) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Sample) GetValue() float64 {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return 0
}

func (m *Sample) GetTimestampMs() int64 {
	if m != nil && m.TimestampMs != nil {
		return *m.TimestampMs
	}
	return 0
}

type LabelPair struct {
	Name             *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Value            *string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *LabelPair) Reset()                    { *m = LabelPair{} }
func (m *LabelPair) String() string            { return proto.CompactTextString(m) }
func (*LabelPair) ProtoMessage()               {}
func (*LabelPair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *LabelPair) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *LabelPair) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

type TimeSeries struct {
	Name   *string      `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Labels []*LabelPair `protobuf:"bytes,2,rep,name=labels" json:"labels,omitempty"`
	// Sorted by time, oldest sample first.
	Samples          []*Sample `protobuf:"bytes,3,rep,name=samples" json:"samples,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *TimeSeries) Reset()                    { *m = TimeSeries{} }
func (m *TimeSeries) String() string            { return proto.CompactTextString(m) }
func (*TimeSeries) ProtoMessage()               {}
func (*TimeSeries) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *TimeSeries) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *TimeSeries) GetLabels() []*LabelPair {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *TimeSeries) GetSamples() []*Sample {
	if m != nil {
		return m.Samples
	}
	return nil
}

type GenericWriteRequest struct {
	Timeseries       []*TimeSeries `protobuf:"bytes,1,rep,name=timeseries" json:"timeseries,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *GenericWriteRequest) Reset()                    { *m = GenericWriteRequest{} }
func (m *GenericWriteRequest) String() string            { return proto.CompactTextString(m) }
func (*GenericWriteRequest) ProtoMessage()               {}
func (*GenericWriteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GenericWriteRequest) GetTimeseries() []*TimeSeries {
	if m != nil {
		return m.Timeseries
	}
	return nil
}

type LabelMatcher struct {
	Type             *MatchType `protobuf:"varint,1,opt,name=type,enum=generic.MatchType" json:"type,omitempty"`
	Name             *string    `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Value            *string    `protobuf:"bytes,3,opt,name=value" json:"value,omitempty"`
	XXX_unrecognized []byte     `json:"-"`
}

func (m *LabelMatcher) Reset()                    { *m = LabelMatcher{} }
func (m *LabelMatcher) String() string            { return proto.CompactTextString(m) }
func (*LabelMatcher) ProtoMessage()               {}
func (*LabelMatcher) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *LabelMatcher) GetType() MatchType {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return MatchType_EQUAL
}

func (m *LabelMatcher) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *LabelMatcher) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

type GenericReadRequest struct {
	StartTimestampMs *int64          `protobuf:"varint,1,opt,name=start_timestamp_ms" json:"start_timestamp_ms,omitempty"`
	EndTimestampMs   *int64          `protobuf:"varint,2,opt,name=end_timestamp_ms" json:"end_timestamp_ms,omitempty"`
	Matchers         []*LabelMatcher `protobuf:"bytes,3,rep,name=matchers" json:"matchers,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *GenericReadRequest) Reset()                    { *m = GenericReadRequest{} }
func (m *GenericReadRequest) String() string            { return proto.CompactTextString(m) }
func (*GenericReadRequest) ProtoMessage()               {}
func (*GenericReadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GenericReadRequest) GetStartTimestampMs() int64 {
	if m != nil && m.StartTimestampMs != nil {
		return *m.StartTimestampMs
	}
	return 0
}

func (m *GenericReadRequest) GetEndTimestampMs() int64 {
	if m != nil && m.EndTimestampMs != nil {
		return *m.EndTimestampMs
	}
	return 0
}

func (m *GenericReadRequest) GetMatchers() []*LabelMatcher {
	if m != nil {
		return m.Matchers
	}
	return nil
}

type GenericReadResponse struct {
	Timeseries       []*TimeSeries `protobuf:"bytes,1,rep,name=timeseries" json:"timeseries,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *GenericReadResponse) Reset()                    { *m = GenericReadResponse{} }
func (m *GenericReadResponse) String() string            { return proto.CompactTextString(m) }
func (*GenericReadResponse) ProtoMessage()               {}
func (*GenericReadResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GenericReadResponse) GetTimeseries() []*TimeSeries {
	if m != nil {
		return m.Timeseries
	}
	return nil
}

func init() {
	proto.RegisterType((*Sample)(nil), "generic.Sample")
	proto.RegisterType((*LabelPair)(nil), "generic.LabelPair")
	proto.RegisterType((*TimeSeries)(nil), "generic.TimeSeries")
	proto.RegisterType((*GenericWriteRequest)(nil), "generic.GenericWriteRequest")
	proto.RegisterType((*LabelMatcher)(nil), "generic.LabelMatcher")
	proto.RegisterType((*GenericReadRequest)(nil), "generic.GenericReadRequest")
	proto.RegisterType((*GenericReadResponse)(nil), "generic.GenericReadResponse")
	proto.RegisterEnum("generic.MatchType", MatchType_name, MatchType_value)
}

var fileDescriptor0 = []byte{
	// 339 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x90, 0x4b, 0x4f, 0xfa, 0x40,
	0x14, 0xc5, 0xff, 0xa5, 0x3c, 0xfe, 0xbd, 0x3c, 0x73, 0xd1, 0xa4, 0x71, 0x45, 0xba, 0x91, 0x98,
	0xc8, 0xc2, 0x0f, 0x60, 0x42, 0x0c, 0xc1, 0x18, 0x1e, 0x0a, 0x35, 0xba, 0xab, 0x23, 0xdc, 0x68,
	0x93, 0xb6, 0xd4, 0xce, 0x60, 0xe2, 0xb7, 0x77, 0x1e, 0x85, 0x82, 0x6e, 0x5c, 0xce, 0x9d, 0x7b,
	0xee, 0xf9, 0x9d, 0x03, 0xcd, 0x37, 0x4a, 0x28, 0x0b, 0x57, 0x83, 0x34, 0xdb, 0x88, 0x0d, 0xd6,
	0xf2, 0xa7, 0x77, 0x09, 0xd5, 0x25, 0x8b, 0xd3, 0x88, 0xb0, 0x09, 0x95, 0x4f, 0x16, 0x6d, 0xc9,
	0xb5, 0x7a, 0x56, 0xdf, 0xc2, 0x13, 0x68, 0x88, 0x30, 0x26, 0x2e, 0xe4, 0x6f, 0x10, 0x73, 0xb7,
	0x24, 0xa7, 0xb6, 0xd7, 0x07, 0x67, 0xc2, 0x5e, 0x29, 0xba, 0x67, 0x61, 0x86, 0x0d, 0x28, 0x27,
	0x2c, 0x36, 0x02, 0xa7, 0xd0, 0xab, 0x4d, 0xc7, 0x7b, 0x01, 0xf0, 0xa5, 0x7e, 0x29, 0x5d, 0x88,
	0xff, 0x58, 0xf5, 0xa0, 0x1a, 0xa9, 0x2b, 0xea, 0xaa, 0xdd, 0xaf, 0x5f, 0xe1, 0x60, 0x47, 0x57,
	0x1c, 0xef, 0x41, 0x8d, 0x6b, 0x30, 0xee, 0xda, 0x7a, 0xa9, 0xbd, 0x5f, 0x32, 0xc0, 0xde, 0x35,
	0x74, 0xc7, 0x66, 0xf2, 0x94, 0x85, 0x82, 0x16, 0xf4, 0xb1, 0x95, 0xb8, 0x78, 0x0e, 0xa0, 0xc1,
	0xb5, 0xb1, 0x34, 0x54, 0xda, 0xee, 0x5e, 0x5b, 0x30, 0x79, 0x53, 0x68, 0x68, 0xbb, 0x29, 0x13,
	0xab, 0x77, 0x52, 0x8e, 0x65, 0xf1, 0x95, 0x1a, 0xc6, 0xd6, 0x01, 0x93, 0xfe, 0xf7, 0xe5, 0xcf,
	0x3e, 0x45, 0xe9, 0x38, 0xb0, 0xad, 0x03, 0x73, 0xc0, 0x1c, 0x67, 0x41, 0x6c, 0xbd, 0xa3, 0x39,
	0x03, 0x94, 0x15, 0x66, 0x22, 0x38, 0x2a, 0x53, 0x59, 0xd8, 0xe8, 0x42, 0x87, 0x92, 0x75, 0xf0,
	0xbb, 0x66, 0x99, 0xe1, 0x7f, 0x6c, 0xa8, 0x76, 0xe9, 0x4f, 0x8f, 0x2b, 0xca, 0x99, 0x0f, 0x3a,
	0x30, 0xa6, 0x3c, 0xdd, 0x24, 0x9c, 0xfe, 0xdc, 0xc1, 0xc5, 0x1d, 0x38, 0x45, 0x3c, 0x07, 0x2a,
	0xa3, 0x87, 0xc7, 0xe1, 0xa4, 0xf3, 0x4f, 0x66, 0x73, 0x66, 0x73, 0x3f, 0x30, 0x4f, 0x0b, 0xdb,
	0x50, 0x5f, 0x8c, 0xc6, 0xa3, 0xe7, 0x60, 0x3a, 0xf4, 0x6f, 0x6e, 0x3b, 0x25, 0x44, 0x68, 0x99,
	0xc1, 0x6c, 0x9e, 0xcf, 0xec, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x23, 0x51, 0xa0, 0x32, 0x63,
	0x02, 0x00, 0x00,
}
