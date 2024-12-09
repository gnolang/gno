package groups

import (
	"gno.land/p/demo/gor/checkpoints"
	"std"
	"strconv"
	"time"
)

type Member struct {
	address           std.Address
	weightCheckPoints *checkpoints.History // maintain snapshot weight instead of snapshot tokens
	metadata          string
	createdAt         time.Time
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t[" + ", " + member.address.String() + ", " + strconv.FormatUint(member.weightCheckPoints.LatestValue(), 10) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}
