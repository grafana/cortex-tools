// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ring.proto

package ring

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_sortkeys "github.com/gogo/protobuf/sortkeys"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strconv "strconv"
	strings "strings"
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

type InstanceState int32

const (
	ACTIVE  InstanceState = 0
	LEAVING InstanceState = 1
	PENDING InstanceState = 2
	JOINING InstanceState = 3
	// This state is only used by gossiping code to distribute information about
	// instances that have been removed from the ring. Ring users should not use it directly.
	LEFT InstanceState = 4
)

var InstanceState_name = map[int32]string{
	0: "ACTIVE",
	1: "LEAVING",
	2: "PENDING",
	3: "JOINING",
	4: "LEFT",
}

var InstanceState_value = map[string]int32{
	"ACTIVE":  0,
	"LEAVING": 1,
	"PENDING": 2,
	"JOINING": 3,
	"LEFT":    4,
}

func (InstanceState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_26381ed67e202a6e, []int{0}
}

// Desc is the top-level type used to model a ring, containing information for individual instances.
type Desc struct {
	Ingesters map[string]InstanceDesc `protobuf:"bytes,1,rep,name=ingesters,proto3" json:"ingesters" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *Desc) Reset()      { *m = Desc{} }
func (*Desc) ProtoMessage() {}
func (*Desc) Descriptor() ([]byte, []int) {
	return fileDescriptor_26381ed67e202a6e, []int{0}
}
func (m *Desc) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Desc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Desc.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Desc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Desc.Merge(m, src)
}
func (m *Desc) XXX_Size() int {
	return m.Size()
}
func (m *Desc) XXX_DiscardUnknown() {
	xxx_messageInfo_Desc.DiscardUnknown(m)
}

var xxx_messageInfo_Desc proto.InternalMessageInfo

func (m *Desc) GetIngesters() map[string]InstanceDesc {
	if m != nil {
		return m.Ingesters
	}
	return nil
}

// InstanceDesc is the top-level type used to model per-instance information in a ring.
type InstanceDesc struct {
	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	// Unix timestamp (with seconds precision) of the last heartbeat sent
	// by this instance.
	Timestamp int64         `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	State     InstanceState `protobuf:"varint,3,opt,name=state,proto3,enum=ring.InstanceState" json:"state,omitempty"`
	Tokens    []uint32      `protobuf:"varint,6,rep,packed,name=tokens,proto3" json:"tokens,omitempty"`
	Zone      string        `protobuf:"bytes,7,opt,name=zone,proto3" json:"zone,omitempty"`
	// Unix timestamp (with seconds precision) of when the instance has been registered
	// to the ring. This field has not been called "joined_timestamp" intentionally, in order
	// to not introduce any misunderstanding with the instance's "joining" state.
	//
	// This field is used to find out subset of instances that could have possibly owned a
	// specific token in the past. Because of this, it's important that either this timestamp
	// is set to the real time the instance has been registered to the ring or it's left
	// 0 (which means unknown).
	//
	// When an instance is already registered in the ring with a value of 0 it's NOT safe to
	// update the timestamp to "now" because it would break the contract, given the instance
	// was already registered before "now". If unknown (0), it should be left as is, and the
	// code will properly deal with that.
	RegisteredTimestamp int64 `protobuf:"varint,8,opt,name=registered_timestamp,json=registeredTimestamp,proto3" json:"registered_timestamp,omitempty"`
}

func (m *InstanceDesc) Reset()      { *m = InstanceDesc{} }
func (*InstanceDesc) ProtoMessage() {}
func (*InstanceDesc) Descriptor() ([]byte, []int) {
	return fileDescriptor_26381ed67e202a6e, []int{1}
}
func (m *InstanceDesc) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InstanceDesc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InstanceDesc.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InstanceDesc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InstanceDesc.Merge(m, src)
}
func (m *InstanceDesc) XXX_Size() int {
	return m.Size()
}
func (m *InstanceDesc) XXX_DiscardUnknown() {
	xxx_messageInfo_InstanceDesc.DiscardUnknown(m)
}

var xxx_messageInfo_InstanceDesc proto.InternalMessageInfo

func (m *InstanceDesc) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *InstanceDesc) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *InstanceDesc) GetState() InstanceState {
	if m != nil {
		return m.State
	}
	return ACTIVE
}

func (m *InstanceDesc) GetTokens() []uint32 {
	if m != nil {
		return m.Tokens
	}
	return nil
}

func (m *InstanceDesc) GetZone() string {
	if m != nil {
		return m.Zone
	}
	return ""
}

func (m *InstanceDesc) GetRegisteredTimestamp() int64 {
	if m != nil {
		return m.RegisteredTimestamp
	}
	return 0
}

func init() {
	proto.RegisterEnum("ring.InstanceState", InstanceState_name, InstanceState_value)
	proto.RegisterType((*Desc)(nil), "ring.Desc")
	proto.RegisterMapType((map[string]InstanceDesc)(nil), "ring.Desc.IngestersEntry")
	proto.RegisterType((*InstanceDesc)(nil), "ring.InstanceDesc")
}

func init() { proto.RegisterFile("ring.proto", fileDescriptor_26381ed67e202a6e) }

var fileDescriptor_26381ed67e202a6e = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x92, 0xc1, 0x6e, 0xd3, 0x40,
	0x18, 0x84, 0xf7, 0x8f, 0x37, 0xae, 0xf3, 0x87, 0x56, 0xd6, 0x16, 0x21, 0x53, 0xa1, 0xc5, 0xea,
	0xc9, 0x20, 0xe1, 0x8a, 0xc0, 0x01, 0x21, 0x71, 0x68, 0xa9, 0x41, 0xb6, 0xa2, 0x50, 0x99, 0xa8,
	0x57, 0xe4, 0x24, 0x8b, 0xb1, 0x4a, 0xec, 0xca, 0xde, 0x20, 0x95, 0x13, 0x8f, 0xc0, 0x0b, 0x70,
	0xe7, 0x51, 0x7a, 0xcc, 0x09, 0xf5, 0x84, 0x88, 0x73, 0xe1, 0xd8, 0x47, 0x40, 0xbb, 0x6e, 0x65,
	0x72, 0x9b, 0xcf, 0x33, 0xff, 0x8c, 0x2d, 0x19, 0xb1, 0xcc, 0xf2, 0xd4, 0x3f, 0x2f, 0x0b, 0x59,
	0x30, 0xaa, 0xf4, 0xde, 0x93, 0x34, 0x93, 0x9f, 0x16, 0x13, 0x7f, 0x5a, 0xcc, 0x0f, 0xd2, 0x22,
	0x2d, 0x0e, 0xb4, 0x39, 0x59, 0x7c, 0xd4, 0xa4, 0x41, 0xab, 0xe6, 0x68, 0xff, 0x07, 0x20, 0x3d,
	0x16, 0xd5, 0x94, 0xbd, 0xc2, 0x5e, 0x96, 0xa7, 0xa2, 0x92, 0xa2, 0xac, 0x1c, 0x70, 0x0d, 0xaf,
	0x3f, 0xb8, 0xef, 0xeb, 0x76, 0x65, 0xfb, 0xe1, 0xad, 0x17, 0xe4, 0xb2, 0xbc, 0x38, 0xa2, 0x97,
	0xbf, 0x1f, 0x92, 0xb8, 0xbd, 0xd8, 0x3b, 0xc1, 0x9d, 0xcd, 0x08, 0xb3, 0xd1, 0x38, 0x13, 0x17,
	0x0e, 0xb8, 0xe0, 0xf5, 0x62, 0x25, 0x99, 0x87, 0xdd, 0x2f, 0xc9, 0xe7, 0x85, 0x70, 0x3a, 0x2e,
	0x78, 0xfd, 0x01, 0x6b, 0xea, 0xc3, 0xbc, 0x92, 0x49, 0x3e, 0x15, 0x6a, 0x26, 0x6e, 0x02, 0x2f,
	0x3b, 0x2f, 0x20, 0xa2, 0x56, 0xc7, 0x36, 0xf6, 0x7f, 0x01, 0xde, 0xf9, 0x3f, 0xc1, 0x18, 0xd2,
	0x64, 0x36, 0x2b, 0x6f, 0x7a, 0xb5, 0x66, 0x0f, 0xb0, 0x27, 0xb3, 0xb9, 0xa8, 0x64, 0x32, 0x3f,
	0xd7, 0xe5, 0x46, 0xdc, 0x3e, 0x60, 0x8f, 0xb0, 0x5b, 0xc9, 0x44, 0x0a, 0xc7, 0x70, 0xc1, 0xdb,
	0x19, 0xec, 0x6e, 0xce, 0xbe, 0x57, 0x56, 0xdc, 0x24, 0xd8, 0x3d, 0x34, 0x65, 0x71, 0x26, 0xf2,
	0xca, 0x31, 0x5d, 0xc3, 0xdb, 0x8e, 0x6f, 0x48, 0x8d, 0x7e, 0x2d, 0x72, 0xe1, 0x6c, 0x35, 0xa3,
	0x4a, 0xb3, 0xa7, 0x78, 0xb7, 0x14, 0x69, 0xa6, 0xbe, 0x58, 0xcc, 0x3e, 0xb4, 0xfb, 0x96, 0xde,
	0xdf, 0x6d, 0xbd, 0xf1, 0xad, 0x15, 0x51, 0x8b, 0xda, 0xdd, 0x88, 0x5a, 0x5d, 0xdb, 0x7c, 0x3c,
	0xc4, 0xed, 0x8d, 0x57, 0x60, 0x88, 0xe6, 0xe1, 0xeb, 0x71, 0x78, 0x1a, 0xd8, 0x84, 0xf5, 0x71,
	0x6b, 0x18, 0x1c, 0x9e, 0x86, 0xa3, 0xb7, 0x36, 0x28, 0x38, 0x09, 0x46, 0xc7, 0x0a, 0x3a, 0x0a,
	0xa2, 0x77, 0xe1, 0x48, 0x81, 0xc1, 0x2c, 0xa4, 0xc3, 0xe0, 0xcd, 0xd8, 0xa6, 0x47, 0xcf, 0x97,
	0x2b, 0x4e, 0xae, 0x56, 0x9c, 0x5c, 0xaf, 0x38, 0x7c, 0xab, 0x39, 0xfc, 0xac, 0x39, 0x5c, 0xd6,
	0x1c, 0x96, 0x35, 0x87, 0x3f, 0x35, 0x87, 0xbf, 0x35, 0x27, 0xd7, 0x35, 0x87, 0xef, 0x6b, 0x4e,
	0x96, 0x6b, 0x4e, 0xae, 0xd6, 0x9c, 0x4c, 0x4c, 0xfd, 0x0f, 0x3c, 0xfb, 0x17, 0x00, 0x00, 0xff,
	0xff, 0x97, 0x76, 0x41, 0xaf, 0x46, 0x02, 0x00, 0x00,
}

func (x InstanceState) String() string {
	s, ok := InstanceState_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (this *Desc) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Desc)
	if !ok {
		that2, ok := that.(Desc)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.Ingesters) != len(that1.Ingesters) {
		return false
	}
	for i := range this.Ingesters {
		a := this.Ingesters[i]
		b := that1.Ingesters[i]
		if !(&a).Equal(&b) {
			return false
		}
	}
	return true
}
func (this *InstanceDesc) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*InstanceDesc)
	if !ok {
		that2, ok := that.(InstanceDesc)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Addr != that1.Addr {
		return false
	}
	if this.Timestamp != that1.Timestamp {
		return false
	}
	if this.State != that1.State {
		return false
	}
	if len(this.Tokens) != len(that1.Tokens) {
		return false
	}
	for i := range this.Tokens {
		if this.Tokens[i] != that1.Tokens[i] {
			return false
		}
	}
	if this.Zone != that1.Zone {
		return false
	}
	if this.RegisteredTimestamp != that1.RegisteredTimestamp {
		return false
	}
	return true
}
func (this *Desc) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&ring.Desc{")
	keysForIngesters := make([]string, 0, len(this.Ingesters))
	for k, _ := range this.Ingesters {
		keysForIngesters = append(keysForIngesters, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForIngesters)
	mapStringForIngesters := "map[string]InstanceDesc{"
	for _, k := range keysForIngesters {
		mapStringForIngesters += fmt.Sprintf("%#v: %#v,", k, this.Ingesters[k])
	}
	mapStringForIngesters += "}"
	if this.Ingesters != nil {
		s = append(s, "Ingesters: "+mapStringForIngesters+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *InstanceDesc) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 10)
	s = append(s, "&ring.InstanceDesc{")
	s = append(s, "Addr: "+fmt.Sprintf("%#v", this.Addr)+",\n")
	s = append(s, "Timestamp: "+fmt.Sprintf("%#v", this.Timestamp)+",\n")
	s = append(s, "State: "+fmt.Sprintf("%#v", this.State)+",\n")
	s = append(s, "Tokens: "+fmt.Sprintf("%#v", this.Tokens)+",\n")
	s = append(s, "Zone: "+fmt.Sprintf("%#v", this.Zone)+",\n")
	s = append(s, "RegisteredTimestamp: "+fmt.Sprintf("%#v", this.RegisteredTimestamp)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringRing(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *Desc) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Desc) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Desc) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Ingesters) > 0 {
		for k := range m.Ingesters {
			v := m.Ingesters[k]
			baseI := i
			{
				size, err := (&v).MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintRing(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintRing(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintRing(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *InstanceDesc) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InstanceDesc) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InstanceDesc) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.RegisteredTimestamp != 0 {
		i = encodeVarintRing(dAtA, i, uint64(m.RegisteredTimestamp))
		i--
		dAtA[i] = 0x40
	}
	if len(m.Zone) > 0 {
		i -= len(m.Zone)
		copy(dAtA[i:], m.Zone)
		i = encodeVarintRing(dAtA, i, uint64(len(m.Zone)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.Tokens) > 0 {
		dAtA3 := make([]byte, len(m.Tokens)*10)
		var j2 int
		for _, num := range m.Tokens {
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		i -= j2
		copy(dAtA[i:], dAtA3[:j2])
		i = encodeVarintRing(dAtA, i, uint64(j2))
		i--
		dAtA[i] = 0x32
	}
	if m.State != 0 {
		i = encodeVarintRing(dAtA, i, uint64(m.State))
		i--
		dAtA[i] = 0x18
	}
	if m.Timestamp != 0 {
		i = encodeVarintRing(dAtA, i, uint64(m.Timestamp))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Addr) > 0 {
		i -= len(m.Addr)
		copy(dAtA[i:], m.Addr)
		i = encodeVarintRing(dAtA, i, uint64(len(m.Addr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintRing(dAtA []byte, offset int, v uint64) int {
	offset -= sovRing(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Desc) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Ingesters) > 0 {
		for k, v := range m.Ingesters {
			_ = k
			_ = v
			l = v.Size()
			mapEntrySize := 1 + len(k) + sovRing(uint64(len(k))) + 1 + l + sovRing(uint64(l))
			n += mapEntrySize + 1 + sovRing(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *InstanceDesc) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Addr)
	if l > 0 {
		n += 1 + l + sovRing(uint64(l))
	}
	if m.Timestamp != 0 {
		n += 1 + sovRing(uint64(m.Timestamp))
	}
	if m.State != 0 {
		n += 1 + sovRing(uint64(m.State))
	}
	if len(m.Tokens) > 0 {
		l = 0
		for _, e := range m.Tokens {
			l += sovRing(uint64(e))
		}
		n += 1 + sovRing(uint64(l)) + l
	}
	l = len(m.Zone)
	if l > 0 {
		n += 1 + l + sovRing(uint64(l))
	}
	if m.RegisteredTimestamp != 0 {
		n += 1 + sovRing(uint64(m.RegisteredTimestamp))
	}
	return n
}

func sovRing(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozRing(x uint64) (n int) {
	return sovRing(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Desc) String() string {
	if this == nil {
		return "nil"
	}
	keysForIngesters := make([]string, 0, len(this.Ingesters))
	for k, _ := range this.Ingesters {
		keysForIngesters = append(keysForIngesters, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForIngesters)
	mapStringForIngesters := "map[string]InstanceDesc{"
	for _, k := range keysForIngesters {
		mapStringForIngesters += fmt.Sprintf("%v: %v,", k, this.Ingesters[k])
	}
	mapStringForIngesters += "}"
	s := strings.Join([]string{`&Desc{`,
		`Ingesters:` + mapStringForIngesters + `,`,
		`}`,
	}, "")
	return s
}
func (this *InstanceDesc) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&InstanceDesc{`,
		`Addr:` + fmt.Sprintf("%v", this.Addr) + `,`,
		`Timestamp:` + fmt.Sprintf("%v", this.Timestamp) + `,`,
		`State:` + fmt.Sprintf("%v", this.State) + `,`,
		`Tokens:` + fmt.Sprintf("%v", this.Tokens) + `,`,
		`Zone:` + fmt.Sprintf("%v", this.Zone) + `,`,
		`RegisteredTimestamp:` + fmt.Sprintf("%v", this.RegisteredTimestamp) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringRing(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Desc) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRing
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
			return fmt.Errorf("proto: Desc: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Desc: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ingesters", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRing
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
				return ErrInvalidLengthRing
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRing
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Ingesters == nil {
				m.Ingesters = make(map[string]InstanceDesc)
			}
			var mapkey string
			mapvalue := &InstanceDesc{}
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowRing
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
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRing
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthRing
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthRing
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRing
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthRing
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return ErrInvalidLengthRing
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &InstanceDesc{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipRing(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthRing
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Ingesters[mapkey] = *mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRing(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRing
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthRing
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
func (m *InstanceDesc) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRing
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
			return fmt.Errorf("proto: InstanceDesc: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InstanceDesc: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Addr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRing
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
				return ErrInvalidLengthRing
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRing
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Addr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRing
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field State", wireType)
			}
			m.State = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRing
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.State |= InstanceState(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType == 0 {
				var v uint32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowRing
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Tokens = append(m.Tokens, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowRing
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
					return ErrInvalidLengthRing
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthRing
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.Tokens) == 0 {
					m.Tokens = make([]uint32, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowRing
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint32(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Tokens = append(m.Tokens, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Tokens", wireType)
			}
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Zone", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRing
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
				return ErrInvalidLengthRing
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRing
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Zone = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RegisteredTimestamp", wireType)
			}
			m.RegisteredTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRing
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RegisteredTimestamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipRing(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRing
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthRing
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
func skipRing(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRing
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
					return 0, ErrIntOverflowRing
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
					return 0, ErrIntOverflowRing
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
				return 0, ErrInvalidLengthRing
			}
			iNdEx += length
			if iNdEx < 0 {
				return 0, ErrInvalidLengthRing
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowRing
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
				next, err := skipRing(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
				if iNdEx < 0 {
					return 0, ErrInvalidLengthRing
				}
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
	ErrInvalidLengthRing = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRing   = fmt.Errorf("proto: integer overflow")
)
