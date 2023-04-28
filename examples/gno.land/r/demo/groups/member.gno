package groups

import (
	"std"
	"strconv"
	"time"
)

type MemberID uint64

type Member struct {
	id        MemberID
	address   std.Address
	weight    int
	metadata  string
	createdAt time.Time
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t\t\t[" + memberIDKey(member.id) + ", " + member.address.String() + ", " + strconv.Itoa(member.weight) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}
