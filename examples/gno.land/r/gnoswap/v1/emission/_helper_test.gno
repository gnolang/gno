package emission

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/avl"

	"gno.land/p/gnoswap/consts"
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
	distributionBpsPct.Set(strconv.Itoa(LIQUIDITY_STAKER), int64(7500))
	distributionBpsPct.Set(strconv.Itoa(DEVOPS), int64(2000))
	distributionBpsPct.Set(strconv.Itoa(COMMUNITY_POOL), int64(500))
	distributionBpsPct.Set(strconv.Itoa(GOV_STAKER), int64(0))

	distributedToStaker = 0
	distributedToDevOps = 0
	distributedToCommunityPool = 0
	distributedToGovStaker = 0
	accuDistributedToStaker = 0
	accuDistributedToDevOps = 0
	accuDistributedToCommunityPool = 0
	accuDistributedToGovStaker = 0

	// Reset emission-specific variables
	leftGNSAmount = 0
	lastExecutedHeight = 0
}
