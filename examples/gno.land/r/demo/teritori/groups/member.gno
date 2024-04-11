package groups

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/teritori/jsonutil"
)

type MemberID uint64

func (mid MemberID) ToJSON() *json.Node {
	return jsonutil.Uint64Node(uint64(mid))
}

func (m *MemberID) FromJSON(ast *json.Node) {
	*m = MemberID(jsonutil.MustUint64(ast))
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

func (member Member) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"id":        member.id.ToJSON(),
		"address":   jsonutil.AddressNode(member.address),
		"weight":    jsonutil.IntNode(member.weight),
		"metadata":  json.StringNode("", member.metadata),
		"createdAt": jsonutil.TimeNode(member.createdAt),
	})
}

func (m *Member) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	m.id.FromJSON(obj["id"])
	m.address = jsonutil.MustAddress(obj["address"])
	m.weight = jsonutil.MustInt(obj["weight"])
	m.metadata = obj["metadata"].MustString()
	m.createdAt = jsonutil.MustTime(obj["createdAt"])
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t\t\t[" + memberIDKey(member.id) + ", " + member.address.String() + ", " + strconv.Itoa(member.weight) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}
