// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: carbonzipper3.proto

package carbonzipperpb3

import (
	encoding_binary "encoding/binary"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
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

type FetchResponseEx struct {
	Name                 string    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	StartTime            int32     `protobuf:"varint,2,opt,name=startTime,proto3" json:"startTime,omitempty"`
	StopTime             int32     `protobuf:"varint,3,opt,name=stopTime,proto3" json:"stopTime,omitempty"`
	StepTime             int32     `protobuf:"varint,4,opt,name=stepTime,proto3" json:"stepTime,omitempty"`
	Values               []float64 `protobuf:"fixed64,5,rep,packed,name=values,proto3" json:"values,omitempty"`
	IsAbsent             []bool    `protobuf:"varint,6,rep,packed,name=isAbsent,proto3" json:"isAbsent,omitempty"`
	RequestedTarget      string    `protobuf:"bytes,7,opt,name=requestedTarget,proto3" json:"requestedTarget,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *FetchResponseEx) Reset()         { *m = FetchResponseEx{} }
func (m *FetchResponseEx) String() string { return proto.CompactTextString(m) }
func (*FetchResponseEx) ProtoMessage()    {}
func (*FetchResponseEx) Descriptor() ([]byte, []int) {
	return fileDescriptor_95f14e18ab2e9726, []int{0}
}
func (m *FetchResponseEx) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FetchResponseEx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FetchResponseEx.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FetchResponseEx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FetchResponseEx.Merge(m, src)
}
func (m *FetchResponseEx) XXX_Size() int {
	return m.Size()
}
func (m *FetchResponseEx) XXX_DiscardUnknown() {
	xxx_messageInfo_FetchResponseEx.DiscardUnknown(m)
}

var xxx_messageInfo_FetchResponseEx proto.InternalMessageInfo

func (m *FetchResponseEx) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FetchResponseEx) GetStartTime() int32 {
	if m != nil {
		return m.StartTime
	}
	return 0
}

func (m *FetchResponseEx) GetStopTime() int32 {
	if m != nil {
		return m.StopTime
	}
	return 0
}

func (m *FetchResponseEx) GetStepTime() int32 {
	if m != nil {
		return m.StepTime
	}
	return 0
}

func (m *FetchResponseEx) GetValues() []float64 {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *FetchResponseEx) GetIsAbsent() []bool {
	if m != nil {
		return m.IsAbsent
	}
	return nil
}

func (m *FetchResponseEx) GetRequestedTarget() string {
	if m != nil {
		return m.RequestedTarget
	}
	return ""
}

type MultiFetchResponseEx struct {
	Metrics              []*FetchResponseEx `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
	Errors               []*Error           `protobuf:"bytes,99,rep,name=errors,proto3" json:"errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *MultiFetchResponseEx) Reset()         { *m = MultiFetchResponseEx{} }
func (m *MultiFetchResponseEx) String() string { return proto.CompactTextString(m) }
func (*MultiFetchResponseEx) ProtoMessage()    {}
func (*MultiFetchResponseEx) Descriptor() ([]byte, []int) {
	return fileDescriptor_95f14e18ab2e9726, []int{1}
}
func (m *MultiFetchResponseEx) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MultiFetchResponseEx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MultiFetchResponseEx.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MultiFetchResponseEx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MultiFetchResponseEx.Merge(m, src)
}
func (m *MultiFetchResponseEx) XXX_Size() int {
	return m.Size()
}
func (m *MultiFetchResponseEx) XXX_DiscardUnknown() {
	xxx_messageInfo_MultiFetchResponseEx.DiscardUnknown(m)
}

var xxx_messageInfo_MultiFetchResponseEx proto.InternalMessageInfo

func (m *MultiFetchResponseEx) GetMetrics() []*FetchResponseEx {
	if m != nil {
		return m.Metrics
	}
	return nil
}

func (m *MultiFetchResponseEx) GetErrors() []*Error {
	if m != nil {
		return m.Errors
	}
	return nil
}

type Error struct {
	Target               string   `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	ErrorMessage         string   `protobuf:"bytes,2,opt,name=errorMessage,proto3" json:"errorMessage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_95f14e18ab2e9726, []int{2}
}
func (m *Error) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Error.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return m.Size()
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

func (m *Error) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *Error) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func init() {
	proto.RegisterType((*FetchResponseEx)(nil), "carbonzipperpb3.FetchResponseEx")
	proto.RegisterType((*MultiFetchResponseEx)(nil), "carbonzipperpb3.MultiFetchResponseEx")
	proto.RegisterType((*Error)(nil), "carbonzipperpb3.Error")
}

func init() { proto.RegisterFile("carbonzipper3.proto", fileDescriptor_95f14e18ab2e9726) }

var fileDescriptor_95f14e18ab2e9726 = []byte{
	// 306 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0xb1, 0x4e, 0xf3, 0x30,
	0x14, 0x85, 0x7f, 0xff, 0x69, 0xd3, 0xf6, 0x82, 0x54, 0x64, 0x50, 0x65, 0x21, 0x14, 0x45, 0x99,
	0x32, 0x65, 0xa0, 0x1b, 0x1b, 0xa0, 0xb2, 0x75, 0xb1, 0xfa, 0x02, 0x4e, 0xb8, 0x2a, 0x91, 0xda,
	0xd8, 0xf8, 0xba, 0x08, 0x31, 0xf2, 0x14, 0x3c, 0x12, 0x23, 0x2b, 0x1b, 0x2a, 0x2f, 0x82, 0x6a,
	0xb7, 0x05, 0xc2, 0xe6, 0x73, 0xbe, 0x7b, 0xa4, 0x7b, 0x8f, 0xe1, 0xb8, 0x52, 0xb6, 0xd4, 0xcd,
	0x53, 0x6d, 0x0c, 0xda, 0x71, 0x61, 0xac, 0x76, 0x9a, 0x0f, 0x7f, 0x9a, 0xa6, 0x1c, 0x67, 0xef,
	0x0c, 0x86, 0x37, 0xe8, 0xaa, 0x3b, 0x89, 0x64, 0x74, 0x43, 0x38, 0x79, 0xe4, 0x1c, 0x3a, 0x8d,
	0x5a, 0xa2, 0x60, 0x29, 0xcb, 0x07, 0xd2, 0xbf, 0xf9, 0x19, 0x0c, 0xc8, 0x29, 0xeb, 0x66, 0xf5,
	0x12, 0xc5, 0xff, 0x94, 0xe5, 0x5d, 0xf9, 0x6d, 0xf0, 0x53, 0xe8, 0x93, 0xd3, 0xc6, 0xc3, 0xc8,
	0xc3, 0xbd, 0x0e, 0x0c, 0x03, 0xeb, 0xec, 0x58, 0xd0, 0x7c, 0x04, 0xf1, 0x83, 0x5a, 0xac, 0x90,
	0x44, 0x37, 0x8d, 0x72, 0x26, 0xb7, 0x6a, 0x93, 0xa9, 0xe9, 0xb2, 0x24, 0x6c, 0x9c, 0x88, 0xd3,
	0x28, 0xef, 0xcb, 0xbd, 0xe6, 0x39, 0x0c, 0x2d, 0xde, 0xaf, 0x90, 0x1c, 0xde, 0xce, 0x94, 0x9d,
	0xa3, 0x13, 0x3d, 0xbf, 0x68, 0xdb, 0xce, 0x9e, 0x19, 0x9c, 0x4c, 0x57, 0x0b, 0x57, 0xb7, 0x0f,
	0xbc, 0x80, 0xde, 0x12, 0x9d, 0xad, 0x2b, 0x12, 0x2c, 0x8d, 0xf2, 0x83, 0xf3, 0xb4, 0x68, 0xf5,
	0x52, 0xb4, 0x22, 0x72, 0x17, 0xe0, 0x05, 0xc4, 0x68, 0xad, 0xb6, 0x24, 0x2a, 0x1f, 0x1d, 0xfd,
	0x89, 0x4e, 0x36, 0x58, 0x6e, 0xa7, 0xb2, 0x6b, 0xe8, 0x7a, 0x63, 0x73, 0xab, 0x0b, 0xeb, 0x86,
	0x5e, 0xb7, 0x8a, 0x67, 0x70, 0xe8, 0x47, 0xa7, 0x48, 0xa4, 0xe6, 0xa1, 0xdc, 0x81, 0xfc, 0xe5,
	0x5d, 0x1d, 0xbd, 0xae, 0x13, 0xf6, 0xb6, 0x4e, 0xd8, 0xc7, 0x3a, 0x61, 0x2f, 0x9f, 0xc9, 0xbf,
	0x32, 0xf6, 0xff, 0x39, 0xfe, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x28, 0x53, 0x5a, 0x16, 0xe6, 0x01,
	0x00, 0x00,
}

func (m *FetchResponseEx) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FetchResponseEx) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FetchResponseEx) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.RequestedTarget) > 0 {
		i -= len(m.RequestedTarget)
		copy(dAtA[i:], m.RequestedTarget)
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.RequestedTarget)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.IsAbsent) > 0 {
		for iNdEx := len(m.IsAbsent) - 1; iNdEx >= 0; iNdEx-- {
			i--
			if m.IsAbsent[iNdEx] {
				dAtA[i] = 1
			} else {
				dAtA[i] = 0
			}
		}
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.IsAbsent)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Values) > 0 {
		for iNdEx := len(m.Values) - 1; iNdEx >= 0; iNdEx-- {
			f1 := math.Float64bits(float64(m.Values[iNdEx]))
			i -= 8
			encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(f1))
		}
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.Values)*8))
		i--
		dAtA[i] = 0x2a
	}
	if m.StepTime != 0 {
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(m.StepTime))
		i--
		dAtA[i] = 0x20
	}
	if m.StopTime != 0 {
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(m.StopTime))
		i--
		dAtA[i] = 0x18
	}
	if m.StartTime != 0 {
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(m.StartTime))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MultiFetchResponseEx) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MultiFetchResponseEx) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MultiFetchResponseEx) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Errors) > 0 {
		for iNdEx := len(m.Errors) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Errors[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintCarbonzipper3(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x6
			i--
			dAtA[i] = 0x9a
		}
	}
	if len(m.Metrics) > 0 {
		for iNdEx := len(m.Metrics) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Metrics[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintCarbonzipper3(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Error) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Error) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Error) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.ErrorMessage) > 0 {
		i -= len(m.ErrorMessage)
		copy(dAtA[i:], m.ErrorMessage)
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.ErrorMessage)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Target) > 0 {
		i -= len(m.Target)
		copy(dAtA[i:], m.Target)
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.Target)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintCarbonzipper3(dAtA []byte, offset int, v uint64) int {
	offset -= sovCarbonzipper3(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *FetchResponseEx) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovCarbonzipper3(uint64(l))
	}
	if m.StartTime != 0 {
		n += 1 + sovCarbonzipper3(uint64(m.StartTime))
	}
	if m.StopTime != 0 {
		n += 1 + sovCarbonzipper3(uint64(m.StopTime))
	}
	if m.StepTime != 0 {
		n += 1 + sovCarbonzipper3(uint64(m.StepTime))
	}
	if len(m.Values) > 0 {
		n += 1 + sovCarbonzipper3(uint64(len(m.Values)*8)) + len(m.Values)*8
	}
	if len(m.IsAbsent) > 0 {
		n += 1 + sovCarbonzipper3(uint64(len(m.IsAbsent))) + len(m.IsAbsent)*1
	}
	l = len(m.RequestedTarget)
	if l > 0 {
		n += 1 + l + sovCarbonzipper3(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *MultiFetchResponseEx) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, e := range m.Metrics {
			l = e.Size()
			n += 1 + l + sovCarbonzipper3(uint64(l))
		}
	}
	if len(m.Errors) > 0 {
		for _, e := range m.Errors {
			l = e.Size()
			n += 2 + l + sovCarbonzipper3(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *Error) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Target)
	if l > 0 {
		n += 1 + l + sovCarbonzipper3(uint64(l))
	}
	l = len(m.ErrorMessage)
	if l > 0 {
		n += 1 + l + sovCarbonzipper3(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovCarbonzipper3(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCarbonzipper3(x uint64) (n int) {
	return sovCarbonzipper3(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FetchResponseEx) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbonzipper3
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: FetchResponseEx: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FetchResponseEx: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartTime", wireType)
			}
			m.StartTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StartTime |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StopTime", wireType)
			}
			m.StopTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StopTime |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StepTime", wireType)
			}
			m.StepTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StepTime |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType == 1 {
				var v uint64
				if (iNdEx + 8) > l {
					return io.ErrUnexpectedEOF
				}
				v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
				iNdEx += 8
				v2 := float64(math.Float64frombits(v))
				m.Values = append(m.Values, v2)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowCarbonzipper3
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthCarbonzipper3
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthCarbonzipper3
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				elementCount = packedLen / 8
				if elementCount != 0 && len(m.Values) == 0 {
					m.Values = make([]float64, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint64
					if (iNdEx + 8) > l {
						return io.ErrUnexpectedEOF
					}
					v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
					iNdEx += 8
					v2 := float64(math.Float64frombits(v))
					m.Values = append(m.Values, v2)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Values", wireType)
			}
		case 6:
			if wireType == 0 {
				var v int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowCarbonzipper3
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.IsAbsent = append(m.IsAbsent, bool(v != 0))
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowCarbonzipper3
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthCarbonzipper3
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthCarbonzipper3
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				elementCount = packedLen
				if elementCount != 0 && len(m.IsAbsent) == 0 {
					m.IsAbsent = make([]bool, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowCarbonzipper3
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.IsAbsent = append(m.IsAbsent, bool(v != 0))
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field IsAbsent", wireType)
			}
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequestedTarget", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RequestedTarget = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbonzipper3(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MultiFetchResponseEx) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbonzipper3
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MultiFetchResponseEx: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MultiFetchResponseEx: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metrics", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metrics = append(m.Metrics, &FetchResponseEx{})
			if err := m.Metrics[len(m.Metrics)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 99:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Errors", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Errors = append(m.Errors, &Error{})
			if err := m.Errors[len(m.Errors)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbonzipper3(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Error) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCarbonzipper3
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Error: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Error: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Target", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Target = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ErrorMessage", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ErrorMessage = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbonzipper3(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCarbonzipper3(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCarbonzipper3
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCarbonzipper3
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthCarbonzipper3
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCarbonzipper3
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCarbonzipper3
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCarbonzipper3        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCarbonzipper3          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCarbonzipper3 = fmt.Errorf("proto: unexpected end of group")
)