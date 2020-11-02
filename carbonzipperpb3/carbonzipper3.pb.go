// Code generated by protoc-gen-gogo.
// source: carbonzipper3.proto
// DO NOT EDIT!

/*
	Package carbonzipperpb3 is a generated protocol buffer package.

	It is generated from these files:
		carbonzipper3.proto

	It has these top-level messages:
		FetchResponse
		MultiFetchResponse
*/
package carbonzipperpb3

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type FetchResponse struct {
	Name            string    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	StartTime       int32     `protobuf:"varint,2,opt,name=startTime,proto3" json:"startTime,omitempty"`
	StopTime        int32     `protobuf:"varint,3,opt,name=stopTime,proto3" json:"stopTime,omitempty"`
	StepTime        int32     `protobuf:"varint,4,opt,name=stepTime,proto3" json:"stepTime,omitempty"`
	Values          []float64 `protobuf:"fixed64,5,rep,packed,name=values" json:"values,omitempty"`
	IsAbsent        []bool    `protobuf:"varint,6,rep,packed,name=isAbsent" json:"isAbsent,omitempty"`
	RequestedTarget string    `protobuf:"bytes,7,opt,name=requestedTarget,proto3" json:"requestedTarget,omitempty"`
}

func (m *FetchResponse) Reset()                    { *m = FetchResponse{} }
func (m *FetchResponse) String() string            { return proto.CompactTextString(m) }
func (*FetchResponse) ProtoMessage()               {}
func (*FetchResponse) Descriptor() ([]byte, []int) { return fileDescriptorCarbonzipper3, []int{0} }

func (m *FetchResponse) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FetchResponse) GetStartTime() int32 {
	if m != nil {
		return m.StartTime
	}
	return 0
}

func (m *FetchResponse) GetStopTime() int32 {
	if m != nil {
		return m.StopTime
	}
	return 0
}

func (m *FetchResponse) GetStepTime() int32 {
	if m != nil {
		return m.StepTime
	}
	return 0
}

func (m *FetchResponse) GetValues() []float64 {
	if m != nil {
		return m.Values
	}
	return nil
}

func (m *FetchResponse) GetIsAbsent() []bool {
	if m != nil {
		return m.IsAbsent
	}
	return nil
}

func (m *FetchResponse) GetRequestedTarget() string {
	if m != nil {
		return m.RequestedTarget
	}
	return ""
}

type MultiFetchResponse struct {
	Metrics []*FetchResponse `protobuf:"bytes,1,rep,name=metrics" json:"metrics,omitempty"`
}

func (m *MultiFetchResponse) Reset()                    { *m = MultiFetchResponse{} }
func (m *MultiFetchResponse) String() string            { return proto.CompactTextString(m) }
func (*MultiFetchResponse) ProtoMessage()               {}
func (*MultiFetchResponse) Descriptor() ([]byte, []int) { return fileDescriptorCarbonzipper3, []int{1} }

func (m *MultiFetchResponse) GetMetrics() []*FetchResponse {
	if m != nil {
		return m.Metrics
	}
	return nil
}

func init() {
	proto.RegisterType((*FetchResponse)(nil), "carbonzipperpb3.FetchResponse")
	proto.RegisterType((*MultiFetchResponse)(nil), "carbonzipperpb3.MultiFetchResponse")
}
func (m *FetchResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FetchResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if m.StartTime != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(m.StartTime))
	}
	if m.StopTime != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(m.StopTime))
	}
	if m.StepTime != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(m.StepTime))
	}
	if len(m.Values) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.Values)*8))
		for _, num := range m.Values {
			f1 := math.Float64bits(float64(num))
			dAtA[i] = uint8(f1)
			i++
			dAtA[i] = uint8(f1 >> 8)
			i++
			dAtA[i] = uint8(f1 >> 16)
			i++
			dAtA[i] = uint8(f1 >> 24)
			i++
			dAtA[i] = uint8(f1 >> 32)
			i++
			dAtA[i] = uint8(f1 >> 40)
			i++
			dAtA[i] = uint8(f1 >> 48)
			i++
			dAtA[i] = uint8(f1 >> 56)
			i++
		}
	}
	if len(m.IsAbsent) > 0 {
		dAtA[i] = 0x32
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.IsAbsent)))
		for _, b := range m.IsAbsent {
			if b {
				dAtA[i] = 1
			} else {
				dAtA[i] = 0
			}
			i++
		}
	}
	if len(m.RequestedTarget) > 0 {
		dAtA[i] = 0x3a
		i++
		i = encodeVarintCarbonzipper3(dAtA, i, uint64(len(m.RequestedTarget)))
		i += copy(dAtA[i:], m.RequestedTarget)
	}
	return i, nil
}

func (m *MultiFetchResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MultiFetchResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, msg := range m.Metrics {
			dAtA[i] = 0xa
			i++
			i = encodeVarintCarbonzipper3(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func encodeFixed64Carbonzipper3(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Carbonzipper3(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintCarbonzipper3(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *FetchResponse) Size() (n int) {
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
	return n
}

func (m *MultiFetchResponse) Size() (n int) {
	var l int
	_ = l
	if len(m.Metrics) > 0 {
		for _, e := range m.Metrics {
			l = e.Size()
			n += 1 + l + sovCarbonzipper3(uint64(l))
		}
	}
	return n
}

func sovCarbonzipper3(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCarbonzipper3(x uint64) (n int) {
	return sovCarbonzipper3(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FetchResponse) Unmarshal(dAtA []byte) error {
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
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: FetchResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FetchResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + intStringLen
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
				m.StartTime |= (int32(b) & 0x7F) << shift
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
				m.StopTime |= (int32(b) & 0x7F) << shift
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
				m.StepTime |= (int32(b) & 0x7F) << shift
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
				iNdEx += 8
				v = uint64(dAtA[iNdEx-8])
				v |= uint64(dAtA[iNdEx-7]) << 8
				v |= uint64(dAtA[iNdEx-6]) << 16
				v |= uint64(dAtA[iNdEx-5]) << 24
				v |= uint64(dAtA[iNdEx-4]) << 32
				v |= uint64(dAtA[iNdEx-3]) << 40
				v |= uint64(dAtA[iNdEx-2]) << 48
				v |= uint64(dAtA[iNdEx-1]) << 56
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
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthCarbonzipper3
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				for iNdEx < postIndex {
					var v uint64
					if (iNdEx + 8) > l {
						return io.ErrUnexpectedEOF
					}
					iNdEx += 8
					v = uint64(dAtA[iNdEx-8])
					v |= uint64(dAtA[iNdEx-7]) << 8
					v |= uint64(dAtA[iNdEx-6]) << 16
					v |= uint64(dAtA[iNdEx-5]) << 24
					v |= uint64(dAtA[iNdEx-4]) << 32
					v |= uint64(dAtA[iNdEx-3]) << 40
					v |= uint64(dAtA[iNdEx-2]) << 48
					v |= uint64(dAtA[iNdEx-1]) << 56
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
					v |= (int(b) & 0x7F) << shift
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
					packedLen |= (int(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthCarbonzipper3
				}
				postIndex := iNdEx + packedLen
				if postIndex > l {
					return io.ErrUnexpectedEOF
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
						v |= (int(b) & 0x7F) << shift
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
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + intStringLen
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
			if skippy < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MultiFetchResponse) Unmarshal(dAtA []byte) error {
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
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MultiFetchResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MultiFetchResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metrics = append(m.Metrics, &FetchResponse{})
			if err := m.Metrics[len(m.Metrics)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCarbonzipper3(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCarbonzipper3
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
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
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthCarbonzipper3
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCarbonzipper3
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipCarbonzipper3(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthCarbonzipper3 = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCarbonzipper3   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("carbonzipper3.proto", fileDescriptorCarbonzipper3) }

var fileDescriptorCarbonzipper3 = []byte{
	// 254 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0x31, 0x4e, 0xc3, 0x30,
	0x18, 0x85, 0x31, 0x69, 0xd3, 0xd6, 0x08, 0x15, 0x19, 0x09, 0x59, 0x08, 0x59, 0x56, 0x27, 0x4f,
	0x19, 0xc8, 0xc2, 0x0a, 0x03, 0x1b, 0x0c, 0x56, 0x2f, 0xe0, 0x84, 0x5f, 0x60, 0xa9, 0x89, 0x8d,
	0xfd, 0x87, 0x81, 0x93, 0x70, 0x24, 0x46, 0x36, 0x56, 0x14, 0x2e, 0x82, 0x70, 0x68, 0x69, 0xbb,
	0xf9, 0x7b, 0x9f, 0x9f, 0x64, 0x3f, 0x7a, 0x5a, 0x9b, 0x50, 0xb9, 0xf6, 0xd5, 0x7a, 0x0f, 0xa1,
	0x2c, 0x7c, 0x70, 0xe8, 0xd8, 0x7c, 0x3b, 0xf4, 0x55, 0xb9, 0xf8, 0x24, 0xf4, 0xf8, 0x16, 0xb0,
	0x7e, 0xd2, 0x10, 0xbd, 0x6b, 0x23, 0x30, 0x46, 0x47, 0xad, 0x69, 0x80, 0x13, 0x49, 0xd4, 0x4c,
	0xa7, 0x33, 0xbb, 0xa0, 0xb3, 0x88, 0x26, 0xe0, 0xd2, 0x36, 0xc0, 0x0f, 0x25, 0x51, 0x63, 0xfd,
	0x1f, 0xb0, 0x73, 0x3a, 0x8d, 0xe8, 0x7c, 0x92, 0x59, 0x92, 0x1b, 0x1e, 0x1c, 0x0c, 0x6e, 0xb4,
	0x76, 0x03, 0xb3, 0x33, 0x9a, 0xbf, 0x98, 0x55, 0x07, 0x91, 0x8f, 0x65, 0xa6, 0x88, 0xfe, 0xa3,
	0xdf, 0x8e, 0x8d, 0xd7, 0x55, 0x84, 0x16, 0x79, 0x2e, 0x33, 0x35, 0xd5, 0x1b, 0x66, 0x8a, 0xce,
	0x03, 0x3c, 0x77, 0x10, 0x11, 0x1e, 0x96, 0x26, 0x3c, 0x02, 0xf2, 0x49, 0x7a, 0xe8, 0x7e, 0xbc,
	0xb8, 0xa7, 0xec, 0xae, 0x5b, 0xa1, 0xdd, 0xfd, 0xdd, 0x15, 0x9d, 0x34, 0x80, 0xc1, 0xd6, 0x91,
	0x13, 0x99, 0xa9, 0xa3, 0x4b, 0x51, 0xec, 0x4d, 0x52, 0xec, 0x14, 0xf4, 0xfa, 0xfa, 0xcd, 0xc9,
	0x7b, 0x2f, 0xc8, 0x47, 0x2f, 0xc8, 0x57, 0x2f, 0xc8, 0xdb, 0xb7, 0x38, 0xa8, 0xf2, 0xb4, 0x69,
	0xf9, 0x13, 0x00, 0x00, 0xff, 0xff, 0x4d, 0x54, 0x25, 0x39, 0x6a, 0x01, 0x00, 0x00,
}
