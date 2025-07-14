// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. bar:qux:100

// POSITIONs:
// 1. in-range
// 2. (will be unstaked) in-range

// REWARDs:
// - external bar ( bar:qux:100 )

package staker_test

import (
	"std"
	"strconv"
	"time"

	"github.com/gnoswap-labs/gnoswap/contract/r/gnoswap/test_token/bar"
	"github.com/gnoswap-labs/gnoswap/contract/r/gnoswap/test_token/qux"

	"gno.land/p/demo/grc/grc721"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
	pl "gno.land/r/gnoswap/v1/pool"
	sr "gno.land/r/gnoswap/v1/staker"

	"gno.land/r/gnoswap/v1/access"

	"gno.land/r/gnoswap/v1/gnft"
	"gno.land/r/gnoswap/v1/gns"

	pn "gno.land/r/gnoswap/v1/position"
)

var (
	adminAddr, _ = access.GetAddress(access.ROLE_ADMIN)
	adminUser    = adminAddr
	adminRealm   = std.NewUserRealm(adminAddr)

	stakerAddr, _ = access.GetAddress(access.ROLE_STAKER)
	poolAddr, _   = access.GetAddress(access.ROLE_POOL)

	barPath = "gno.land/r/onbloc/bar"
	quxPath = "gno.land/r/onbloc/qux"

	fee100      uint32 = 100
	max_timeout int64  = 9999999999

	// external incentive deposit fee
	depositGnsAmount uint64 = 1_000_000_000 // 1_000 GNS
	TIMESTAMP_90DAYS int64  = 90 * 24 * 60 * 60
)

func main() {
	testInit()
	testCreatePool()
	testMintBarQuxPos01()
	testMintBarQuxPos02()

	testCreateExternalIncentive()

	testStakeTokenPos01()
	testStakeTokenPos02()
	testMakeExternalBarStart()

	testCollectReward()
	testUnstakeTokenPos02()
	testCollectRewardAfterUnstake()
}

func testInit() {
	testing.SetRealm(adminRealm)

	sr.SetUnStakingFeeByAdmin(0)
}

func testCreatePool() {
	testing.SetRealm(adminRealm)

	cross(pl.SetPoolCreationFeeByAdmin)(0)

	testing.SkipHeights(1)
	pl.CreatePool(
		barPath,
		quxPath,
		fee100,
		common.TickMathGetSqrtRatioAtTick(0).ToString(), // 79228162514264337593543950337
	)
}

func testMintBarQuxPos01() {
	testing.SetRealm(adminRealm)

	bar.Approve(poolAddr, consts.UINT64_MAX)
	qux.Approve(poolAddr, consts.UINT64_MAX)

	testing.SkipHeights(1)
	cross(pn.Mint)(
		barPath,
		quxPath,
		fee100,
		int32(-50),
		int32(50),
		"50",
		"50",
		"1",
		"1",
		max_timeout,
		adminAddr,
		adminAddr,
	)
}

func testMintBarQuxPos02() {
	testing.SetRealm(adminRealm)

	bar.Approve(poolAddr, consts.UINT64_MAX)
	qux.Approve(poolAddr, consts.UINT64_MAX)

	testing.SkipHeights(1)
	cross(pn.Mint)(
		barPath,
		quxPath,
		fee100,
		int32(-50),
		int32(50),
		"50",
		"50",
		"1",
		"1",
		max_timeout,
		adminAddr,
		adminAddr,
	)
}

func testCreateExternalIncentive() {
	testing.SetRealm(adminRealm)

	bar.Approve(stakerAddr, consts.UINT64_MAX)
	gns.Approve(stakerAddr, depositGnsAmount)

	testing.SkipHeights(1)
	sr.CreateExternalIncentive(
		"gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100",
		barPath,
		9000000000,
		1234569600,
		1234569600+TIMESTAMP_90DAYS,
	)
}

func testStakeTokenPos01() {
	testing.SetRealm(adminRealm)

	testing.SkipHeights(1)

	gnft.Approve(stakerAddr, positionIdFrom(1))
	sr.StakeToken(1, "")
}

func testStakeTokenPos02() {
	testing.SetRealm(adminRealm)

	gnft.Approve(stakerAddr, positionIdFrom(2))
	sr.StakeToken(2, "")
}

func testMakeExternalBarStart() {
	externalStartTime := int64(1234569600)
	nowTime := time.Now().Unix()

	timeLeft := externalStartTime - nowTime

	blockAvgTime := milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockLeft := timeLeft / blockAvgTime
	testing.SkipHeights(int64(blockLeft)) // skip until external bar starts
	testing.SkipHeights(10)               // skip bit more to see reward calculation

	// check reward for position 01 (in-range)
	testing.SetRealm(adminRealm)
	beforeBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	afterBar := bar.BalanceOf(adminUser)
	diff := afterBar - beforeBar
	if diff == 0 {
		panic("position 01 in-range, should have reward")
	}
}

func testCollectReward() {
	testing.SetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	testing.SkipHeights(1)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar
	if !isInErrorRange(diff, 347) { // 2314(rewardPerBlock) * 50%(position01Ratio) * 30%(warmUp)
		panic("expected about 347")
	}
}

func testUnstakeTokenPos02() {
	testing.SetRealm(adminRealm)
	sr.UnStakeToken(2, false) // position 02 is unstaked, position 01 the only staked position
}

func testCollectRewardAfterUnstake() {
	testing.SetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	testing.SkipHeights(1)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar
	if !isInErrorRange(diff, 693) { // 2314(rewardPerBlock) * 30%(warmUp)
		panic("expected about 693")
	}
}

func positionIdFrom(positionId any) grc721.TokenID {
	if positionId == nil {
		panic("positionId is nil")
	}

	switch positionId.(type) {
	case string:
		return grc721.TokenID(positionId.(string))
	case int:
		return grc721.TokenID(strconv.Itoa(positionId.(int)))
	case uint64:
		return grc721.TokenID(strconv.Itoa(int(positionId.(uint64))))
	case grc721.TokenID:
		return positionId.(grc721.TokenID)
	default:
		panic("unsupported positionId type")
	}
}

func milliToSec(ms int64) int64 {
	return ms / consts.MILLISECONDS_PER_SECOND
}
