package groups

import (
	"gno.land/p/demo/gor/checkpoints"
	"std"
	"strconv"
	"time"
)

type MemberID uint64

type Member struct {
	id                MemberID
	address           std.Address
	weightCheckPoints *checkpoints.History // maintain snapshot weight instead of snapshot tokens
	metadata          string
	createdAt         time.Time
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t[" + memberIDKey(member.id) + ", " + member.address.String() + ", " + strconv.FormatUint(member.weightCheckPoints.LatestValue(), 10) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}
