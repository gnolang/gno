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

	"gno.land/p/demo/grc/grc721"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	_ "gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/qux"

	"gno.land/r/gnoswap/v1/gnft"

	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"
	sr "gno.land/r/gnoswap/v1/staker"
)

var (
	adminAddr  = consts.ADMIN
	adminUser  = adminAddr
	adminRealm = std.NewUserRealm(adminAddr)

	stakerAddr  = consts.STAKER_ADDR
	stakerUser  = stakerAddr
	stakerRealm = std.NewCodeRealm(consts.STAKER_PATH)

	fooPath = "gno.land/r/onbloc/foo"
	barPath = "gno.land/r/onbloc/bar"
	bazPath = "gno.land/r/onbloc/baz"
	quxPath = "gno.land/r/onbloc/qux"
	oblPath = "gno.land/r/onbloc/obl"

	gnsPath    = "gno.land/r/gnoswap/v1/gns"
	wugnotPath = "gno.land/r/demo/wugnot"

	fee100  uint32 = 100
	fee500  uint32 = 500
	fee3000 uint32 = 3000

	max_timeout int64 = 9999999999

	// external incentive deposit fee
	depositGnsAmount uint64 = 1_000_000_000 // 1_000 GNS

	TIMESTAMP_90DAYS  int64 = 90 * 24 * 60 * 60
	TIMESTAMP_180DAYS int64 = 180 * 24 * 60 * 60
	TIMESTAMP_365DAYS int64 = 365 * 24 * 60 * 60

	poolPath = "gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100"
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
	std.TestSetRealm(adminRealm)

	sr.SetUnStakingFeeByAdmin(0)
}

func testCreatePool() {
	std.TestSetRealm(adminRealm)

	pl.SetPoolCreationFeeByAdmin(0)

	std.TestSkipHeights(1)
	pl.CreatePool(
		barPath,
		quxPath,
		fee100,
		common.TickMathGetSqrtRatioAtTick(0).ToString(), // 79228162514264337593543950337
	)
}

func testMintBarQuxPos01() {
	std.TestSetRealm(adminRealm)

	bar.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	qux.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	std.TestSkipHeights(1)
	pn.Mint(
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
	std.TestSetRealm(adminRealm)

	bar.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	qux.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	std.TestSkipHeights(1)
	pn.Mint(
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
	std.TestSetRealm(adminRealm)

	bar.Approve(consts.STAKER_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.STAKER_ADDR, depositGnsAmount)

	std.TestSkipHeights(1)
	sr.CreateExternalIncentive(
		"gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100",
		barPath,
		9000000000,
		1234569600,
		1234569600+TIMESTAMP_90DAYS,
	)
}

func testStakeTokenPos01() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(1)

	gnft.Approve(stakerAddr, tokenIdFrom(1))
	sr.StakeToken(1)
}

func testStakeTokenPos02() {
	std.TestSetRealm(adminRealm)

	gnft.Approve(stakerAddr, tokenIdFrom(2))
	sr.StakeToken(2)
}

func testMakeExternalBarStart() {
	externalStartTime := int64(1234569600)
	nowTime := time.Now().Unix()

	timeLeft := externalStartTime - nowTime

	blockAvgTime := milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockLeft := timeLeft / blockAvgTime
	std.TestSkipHeights(int64(blockLeft)) // skip until external bar starts
	std.TestSkipHeights(10)               // skip bit more to see reward calculation

	// check reward for position 01 (in-range)
	std.TestSetRealm(adminRealm)
	beforeBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	afterBar := bar.BalanceOf(adminUser)
	diff := afterBar - beforeBar
	if diff == 0 {
		panic("position 01 in-range, should have reward")
	}
}

func testCollectReward() {
	std.TestSetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	std.TestSkipHeights(1)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar
	if !isInErrorRange(diff, 347) { // 2314(rewardPerBlock) * 50%(position01Ratio) * 30%(warmUp)
		panic("expected about 347")
	}
}

func testUnstakeTokenPos02() {
	std.TestSetRealm(adminRealm)
	sr.UnStakeToken(2, false) // position 02 is unstaked, position 01 the only staked position
}

func testCollectRewardAfterUnstake() {
	std.TestSetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	std.TestSkipHeights(1)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar
	if !isInErrorRange(diff, 693) { // 2314(rewardPerBlock) * 30%(warmUp)
		panic("expected about 693")
	}
}

func tokenIdFrom(tokenId interface{}) grc721.TokenID {
	if tokenId == nil {
		panic("tokenId is nil")
	}

	switch tokenId.(type) {
	case string:
		return grc721.TokenID(tokenId.(string))
	case int:
		return grc721.TokenID(strconv.Itoa(tokenId.(int)))
	case uint64:
		return grc721.TokenID(strconv.Itoa(int(tokenId.(uint64))))
	case grc721.TokenID:
		return tokenId.(grc721.TokenID)
	default:
		panic("unsupported tokenId type")
	}
}

func isInErrorRange(expected uint64, actual uint64) bool {
	lowerBound := expected * 999999 / 1000000
	upperBound := expected * 1000001 / 1000000
	return actual >= lowerBound && actual <= upperBound
}

func milliToSec(ms int64) int64 {
	return ms / consts.MILLISECONDS_PER_SECOND
}
