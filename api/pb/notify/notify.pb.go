// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: notify.proto

package notify

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type SendRequest_MsgType int32

const (
	SendRequest_MARKETING SendRequest_MsgType = 0
	SendRequest_BLOCK     SendRequest_MsgType = 1
)

var SendRequest_MsgType_name = map[int32]string{
	0: "MARKETING",
	1: "BLOCK",
}

var SendRequest_MsgType_value = map[string]int32{
	"MARKETING": 0,
	"BLOCK":     1,
}

func (x SendRequest_MsgType) String() string {
	return proto.EnumName(SendRequest_MsgType_name, int32(x))
}

func (SendRequest_MsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_aba76cc4ebe272d4, []int{0, 0}
}

type SendRequest struct {
	Uid   string              `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Type  SendRequest_MsgType `protobuf:"varint,2,opt,name=type,proto3,enum=ataas.notify.SendRequest_MsgType" json:"type,omitempty"`
	Title string              `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	Body  string              `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *SendRequest) Reset()         { *m = SendRequest{} }
func (m *SendRequest) String() string { return proto.CompactTextString(m) }
func (*SendRequest) ProtoMessage()    {}
func (*SendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aba76cc4ebe272d4, []int{0}
}
func (m *SendRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SendRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendRequest.Merge(m, src)
}
func (m *SendRequest) XXX_Size() int {
	return m.Size()
}
func (m *SendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SendRequest proto.InternalMessageInfo

func (m *SendRequest) GetUid() string {
	if m != nil {
		return m.Uid
	}
	return ""
}

func (m *SendRequest) GetType() SendRequest_MsgType {
	if m != nil {
		return m.Type
	}
	return SendRequest_MARKETING
}

func (m *SendRequest) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *SendRequest) GetBody() string {
	if m != nil {
		return m.Body
	}
	return ""
}

type SendResponse struct {
}

func (m *SendResponse) Reset()         { *m = SendResponse{} }
func (m *SendResponse) String() string { return proto.CompactTextString(m) }
func (*SendResponse) ProtoMessage()    {}
func (*SendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aba76cc4ebe272d4, []int{1}
}
func (m *SendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendResponse.Merge(m, src)
}
func (m *SendResponse) XXX_Size() int {
	return m.Size()
}
func (m *SendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SendResponse proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("ataas.notify.SendRequest_MsgType", SendRequest_MsgType_name, SendRequest_MsgType_value)
	proto.RegisterType((*SendRequest)(nil), "ataas.notify.SendRequest")
	proto.RegisterType((*SendResponse)(nil), "ataas.notify.SendResponse")
}

func init() { proto.RegisterFile("notify.proto", fileDescriptor_aba76cc4ebe272d4) }

var fileDescriptor_aba76cc4ebe272d4 = []byte{
	// 286 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xc9, 0xcb, 0x2f, 0xc9,
	0x4c, 0xab, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x49, 0x2c, 0x49, 0x4c, 0x2c, 0xd6,
	0x83, 0x88, 0x29, 0x2d, 0x65, 0xe4, 0xe2, 0x0e, 0x4e, 0xcd, 0x4b, 0x09, 0x4a, 0x2d, 0x2c, 0x4d,
	0x2d, 0x2e, 0x11, 0x12, 0xe0, 0x62, 0x2e, 0xcd, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c,
	0x02, 0x31, 0x85, 0x4c, 0xb9, 0x58, 0x4a, 0x2a, 0x0b, 0x52, 0x25, 0x98, 0x14, 0x18, 0x35, 0xf8,
	0x8c, 0x14, 0xf5, 0x90, 0xb5, 0xeb, 0x21, 0x69, 0xd5, 0xf3, 0x2d, 0x4e, 0x0f, 0xa9, 0x2c, 0x48,
	0x0d, 0x02, 0x2b, 0x17, 0x12, 0xe1, 0x62, 0x2d, 0xc9, 0x2c, 0xc9, 0x49, 0x95, 0x60, 0x06, 0x1b,
	0x05, 0xe1, 0x08, 0x09, 0x71, 0xb1, 0x24, 0xe5, 0xa7, 0x54, 0x4a, 0xb0, 0x80, 0x05, 0xc1, 0x6c,
	0x25, 0x65, 0x2e, 0x76, 0xa8, 0x56, 0x21, 0x5e, 0x2e, 0x4e, 0x5f, 0xc7, 0x20, 0x6f, 0xd7, 0x10,
	0x4f, 0x3f, 0x77, 0x01, 0x06, 0x21, 0x4e, 0x2e, 0x56, 0x27, 0x1f, 0x7f, 0x67, 0x6f, 0x01, 0x46,
	0x25, 0x3e, 0x2e, 0x1e, 0x88, 0x5d, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x46, 0x7e, 0x5c, 0xbc,
	0x7e, 0x60, 0x27, 0x04, 0xa7, 0x16, 0x95, 0x65, 0x26, 0xa7, 0x0a, 0xd9, 0x72, 0xb1, 0x80, 0x14,
	0x08, 0x49, 0xe2, 0x74, 0xa0, 0x94, 0x14, 0x36, 0x29, 0x88, 0x79, 0x4e, 0xce, 0x27, 0x1e, 0xc9,
	0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3, 0x84, 0xc7, 0x72, 0x0c, 0x17, 0x1e,
	0xcb, 0x31, 0xdc, 0x78, 0x2c, 0xc7, 0x10, 0xa5, 0x59, 0x90, 0xab, 0x57, 0x92, 0x9c, 0x56, 0xae,
	0x97, 0x9c, 0x9f, 0xab, 0x97, 0x58, 0xaa, 0x5f, 0x9c, 0x5f, 0x5a, 0x94, 0x9c, 0xaa, 0x0f, 0x36,
	0x4a, 0x3f, 0xb1, 0x20, 0x53, 0xbf, 0x20, 0x49, 0x1f, 0x62, 0x62, 0x12, 0x1b, 0x38, 0x84, 0x8d,
	0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x68, 0x99, 0x4b, 0x8e, 0x71, 0x01, 0x00, 0x00,
}

func (m *SendRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SendRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SendRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Body) > 0 {
		i -= len(m.Body)
		copy(dAtA[i:], m.Body)
		i = encodeVarintNotify(dAtA, i, uint64(len(m.Body)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintNotify(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Type != 0 {
		i = encodeVarintNotify(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Uid) > 0 {
		i -= len(m.Uid)
		copy(dAtA[i:], m.Uid)
		i = encodeVarintNotify(dAtA, i, uint64(len(m.Uid)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintNotify(dAtA []byte, offset int, v uint64) int {
	offset -= sovNotify(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SendRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Uid)
	if l > 0 {
		n += 1 + l + sovNotify(uint64(l))
	}
	if m.Type != 0 {
		n += 1 + sovNotify(uint64(m.Type))
	}
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovNotify(uint64(l))
	}
	l = len(m.Body)
	if l > 0 {
		n += 1 + l + sovNotify(uint64(l))
	}
	return n
}

func (m *SendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovNotify(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozNotify(x uint64) (n int) {
	return sovNotify(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SendRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNotify
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
			return fmt.Errorf("proto: SendRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SendRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uid", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotify
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
				return ErrInvalidLengthNotify
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNotify
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Uid = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotify
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= SendRequest_MsgType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotify
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
				return ErrInvalidLengthNotify
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNotify
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Body", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowNotify
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
				return ErrInvalidLengthNotify
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthNotify
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Body = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipNotify(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNotify
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthNotify
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
func (m *SendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowNotify
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
			return fmt.Errorf("proto: SendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipNotify(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthNotify
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthNotify
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
func skipNotify(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowNotify
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
					return 0, ErrIntOverflowNotify
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
					return 0, ErrIntOverflowNotify
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
				return 0, ErrInvalidLengthNotify
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupNotify
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthNotify
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthNotify        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowNotify          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupNotify = fmt.Errorf("proto: unexpected end of group")
)