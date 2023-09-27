package groups

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/teritori/ujson"
)

type MemberID uint64

func (m *MemberID) FromJSON(ast *ujson.JSONASTNode) {
	var val uint64
	ast.ParseAny(&val)
	*m = MemberID(val)
}

func (mid MemberID) ToJSON() string {
	return ujson.FormatAny(uint64(mid))
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
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "id", Value: member.id},
		{Key: "address", Value: member.address},
		{Key: "weight", Value: member.weight},
		{Key: "metadata", Value: member.metadata},
		{Key: "createdAt", Value: member.createdAt},
	})
}

func (m *Member) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "id", Value: &m.id},
		{Key: "address", Value: &m.address},
		{Key: "weight", Value: &m.weight},
		{Key: "metadata", Value: &m.metadata},
		{Key: "createdAt", Value: &m.createdAt},
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
