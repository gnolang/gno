package emission

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/avl"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
)

var (
	adminRealm     = std.NewUserRealm(adminAddr)
	stakerRealm    = std.NewCodeRealm(consts.STAKER_PATH)
	govRealm       = std.NewCodeRealm(consts.GOV_GOVERNANCE_PATH)
	govStakerRealm = std.NewCodeRealm(consts.GOV_STAKER_PATH)
)

func resetObject(t *testing.T) {
	t.Helper()

	distributionBpsPct = avl.NewTree()
	distributionBpsPct.Set(strconv.Itoa(LIQUIDITY_STAKER), uint64(7500))
	distributionBpsPct.Set(strconv.Itoa(DEVOPS), uint64(2000))
	distributionBpsPct.Set(strconv.Itoa(COMMUNITY_POOL), uint64(500))
	distributionBpsPct.Set(strconv.Itoa(GOV_STAKER), uint64(0))

	distributedToStaker = 0
	distributedToDevOps = 0
	distributedToCommunityPool = 0
	distributedToGovStaker = 0
	accuDistributedToStaker = 0
	accuDistributedToDevOps = 0
	accuDistributedToCommunityPool = 0
	accuDistributedToGovStaker = 0
}

func gnsBalance(t *testing.T, addr std.Address) uint64 {
	t.Helper()

	return gns.BalanceOf(addr)
}
