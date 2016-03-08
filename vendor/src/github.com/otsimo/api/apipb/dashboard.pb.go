// Code generated by protoc-gen-gogo.
// source: dashboard.proto
// DO NOT EDIT!

package apipb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ChartType int32

const (
	ChartType_LINE     ChartType = 0
	ChartType_BAR      ChartType = 1
	ChartType_PIE      ChartType = 2
	ChartType_SCATTER  ChartType = 3
	ChartType_BUBLE    ChartType = 4
	ChartType_RADAR    ChartType = 5
	ChartType_GEO      ChartType = 6
	ChartType_TIMELINE ChartType = 7
)

var ChartType_name = map[int32]string{
	0: "LINE",
	1: "BAR",
	2: "PIE",
	3: "SCATTER",
	4: "BUBLE",
	5: "RADAR",
	6: "GEO",
	7: "TIMELINE",
}
var ChartType_value = map[string]int32{
	"LINE":     0,
	"BAR":      1,
	"PIE":      2,
	"SCATTER":  3,
	"BUBLE":    4,
	"RADAR":    5,
	"GEO":      6,
	"TIMELINE": 7,
}

func (x ChartType) String() string {
	return proto.EnumName(ChartType_name, int32(x))
}

type Dashboard struct {
}

func (m *Dashboard) Reset()         { *m = Dashboard{} }
func (m *Dashboard) String() string { return proto.CompactTextString(m) }
func (*Dashboard) ProtoMessage()    {}

type DashboardGetRequest struct {
	ProfileId  string `protobuf:"bytes,1,opt,name=profile_id,proto3" json:"profile_id,omitempty"`
	ChildId    string `protobuf:"bytes,2,opt,name=child_id,proto3" json:"child_id,omitempty"`
	AppVersion string `protobuf:"bytes,3,opt,name=app_version,proto3" json:"app_version,omitempty"`
}

func (m *DashboardGetRequest) Reset()         { *m = DashboardGetRequest{} }
func (m *DashboardGetRequest) String() string { return proto.CompactTextString(m) }
func (*DashboardGetRequest) ProtoMessage()    {}

func init() {
	proto.RegisterEnum("apipb.ChartType", ChartType_name, ChartType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Client API for DashboardService service

type DashboardServiceClient interface {
	Get(ctx context.Context, in *DashboardGetRequest, opts ...grpc.CallOption) (*Dashboard, error)
}

type dashboardServiceClient struct {
	cc *grpc.ClientConn
}

func NewDashboardServiceClient(cc *grpc.ClientConn) DashboardServiceClient {
	return &dashboardServiceClient{cc}
}

func (c *dashboardServiceClient) Get(ctx context.Context, in *DashboardGetRequest, opts ...grpc.CallOption) (*Dashboard, error) {
	out := new(Dashboard)
	err := grpc.Invoke(ctx, "/apipb.DashboardService/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for DashboardService service

type DashboardServiceServer interface {
	Get(context.Context, *DashboardGetRequest) (*Dashboard, error)
}

func RegisterDashboardServiceServer(s *grpc.Server, srv DashboardServiceServer) {
	s.RegisterService(&_DashboardService_serviceDesc, srv)
}

func _DashboardService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error) {
	in := new(DashboardGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	out, err := srv.(DashboardServiceServer).Get(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _DashboardService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "apipb.DashboardService",
	HandlerType: (*DashboardServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _DashboardService_Get_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

func (m *Dashboard) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *Dashboard) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *DashboardGetRequest) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *DashboardGetRequest) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ProfileId) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintDashboard(data, i, uint64(len(m.ProfileId)))
		i += copy(data[i:], m.ProfileId)
	}
	if len(m.ChildId) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintDashboard(data, i, uint64(len(m.ChildId)))
		i += copy(data[i:], m.ChildId)
	}
	if len(m.AppVersion) > 0 {
		data[i] = 0x1a
		i++
		i = encodeVarintDashboard(data, i, uint64(len(m.AppVersion)))
		i += copy(data[i:], m.AppVersion)
	}
	return i, nil
}

func encodeFixed64Dashboard(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Dashboard(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintDashboard(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *Dashboard) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *DashboardGetRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.ProfileId)
	if l > 0 {
		n += 1 + l + sovDashboard(uint64(l))
	}
	l = len(m.ChildId)
	if l > 0 {
		n += 1 + l + sovDashboard(uint64(l))
	}
	l = len(m.AppVersion)
	if l > 0 {
		n += 1 + l + sovDashboard(uint64(l))
	}
	return n
}

func sovDashboard(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozDashboard(x uint64) (n int) {
	return sovDashboard(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Dashboard) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDashboard
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Dashboard: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Dashboard: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipDashboard(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDashboard
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
func (m *DashboardGetRequest) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDashboard
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DashboardGetRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DashboardGetRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProfileId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDashboard
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDashboard
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ProfileId = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChildId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDashboard
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDashboard
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChildId = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDashboard
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDashboard
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AppVersion = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDashboard(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthDashboard
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
func skipDashboard(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDashboard
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
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
					return 0, ErrIntOverflowDashboard
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
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
					return 0, ErrIntOverflowDashboard
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthDashboard
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowDashboard
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
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
				next, err := skipDashboard(data[start:])
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
	ErrInvalidLengthDashboard = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDashboard   = fmt.Errorf("proto: integer overflow")
)
