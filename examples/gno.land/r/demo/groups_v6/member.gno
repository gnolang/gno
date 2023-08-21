package groups

import (
	"encoding/binary"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/binutils"
	"gno.land/p/demo/jsonutil_v3"
)

type MemberID uint64

func (m *MemberID) FromJSON(ast *jsonutil.JSONASTNode) {
	if ast.ValueKind != jsonutil.JSONTokenKindNumber {
		panic("invalid type")
	}
	*m = MemberID(jsonutil.ParseUint64(ast.Value))
}

func (mid MemberID) ToJSON() string {
	return jsonutil.FormatUint64(uint64(mid))
}

type Member struct {
	id        MemberID
	address   std.Address
	weight    int
	metadata  string
	createdAt time.Time
}

func (m *Member) GetID() MemberID {
	return m.id
}

func (member Member) ToJSON() string {
	return jsonutil.FormatObject([]jsonutil.KeyValue{
		{Key: "id", Value: member.id},
		{Key: "address", Value: member.address.String()},
		{Key: "weight", Value: member.weight},
		{Key: "metadata", Value: member.metadata},
		// {Key: "createdAt", Value: member.createdAt.String()},
	})
}

func (m *Member) FromJSON(ast *jsonutil.JSONASTNode) {
	jsonutil.ParseObjectAST(ast, []*jsonutil.ParseKV{
		{Key: "id", Value: &m.id},
		{Key: "address", Value: &m.address},
		{Key: "weight", Value: &m.weight},
		{Key: "metadata", Value: &m.metadata},
		// {Key: "createdAt", Value: &m.createdAt},
	})
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t\t\t[" + memberIDKey(member.id) + ", " + member.address.String() + ", " + strconv.Itoa(member.weight) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}

func (member *Member) Bytes() []byte {
	b := []byte{}
	b = binary.BigEndian.AppendUint64(b, uint64(member.id))
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(member.address.String())...)
	b = binary.BigEndian.AppendUint32(b, uint32(member.weight))
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(member.metadata)...)
	b = binary.BigEndian.AppendUint64(b, uint64(member.createdAt.UnixNano()))
	return b
}
