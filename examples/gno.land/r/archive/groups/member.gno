package groups

import (
	"strconv"
	"time"
)

type MemberID uint64

type Member struct {
	id          MemberID
	address_XXX address
	weight      int
	metadata    string
	createdAt   time.Time
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t\t\t[" + memberIDKey(member.id) + ", " + member.address_XXX.String() + ", " + strconv.Itoa(member.weight) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}
