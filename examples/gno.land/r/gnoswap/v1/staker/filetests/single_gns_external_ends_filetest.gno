// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. bar:qux:100

// POSITIONs:
// 1. in-range

// REWARDs:
// - external gns 90 days ( bar:qux:100 )

package staker_test

import (
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/testutils"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/gnoswap/v1/access"

	"gno.land/r/gnoswap/v1/gnft"
	"gno.land/r/gnoswap/v1/gns"

	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"
	sr "gno.land/r/gnoswap/v1/staker"
)

var (
	adminAddr, _ = access.GetAddress(access.ROLE_ADMIN)
	adminUser    = adminAddr
	adminRealm   = std.NewUserRealm(adminAddr)

	// g1v4u8getjdeskcsmjv4shgmmjta047h6lua7mup
	externalCreatorAddr  = testutils.TestAddress("externalCreator")
	externalCreatorUser  = externalCreatorAddr
	externalCreatorRealm = std.NewUserRealm(externalCreatorAddr)

	stakerAddr, _ = access.GetAddress(access.ROLE_STAKER)
	poolAddr, _   = access.GetAddress(access.ROLE_POOL)

	barPath = "gno.land/r/onbloc/bar"
	quxPath = "gno.land/r/onbloc/qux"

	gnsPath = "gno.land/r/gnoswap/v1/gns"

	fee100 uint32 = 100

	max_timeout int64 = 9999999999

	// external incentive deposit fee
	depositGnsAmount uint64 = 1_000_000_000 // 1_000 GNS

	TIMESTAMP_90DAYS int64 = 90 * 24 * 60 * 60
)

func main() {
	testInit()
	testCreatePool()
	testMintBarQuxPos01()
	testMintBarQuxPos02()

	testCreateExternalIncentiveGns()
	testMakeExternalStart()

	testStakeTokenPos01()
	testStakeTokenPos02()

	testCollectRewardPos01AndPos02()

	testEndExternalGns()
	testCollectRewardPos01AndPos02AfterEnd()
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

	bar.Approve(poolAddr, consts.UINT64_MAX)
	qux.Approve(poolAddr, consts.UINT64_MAX)

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

	bar.Approve(poolAddr, consts.UINT64_MAX)
	qux.Approve(poolAddr, consts.UINT64_MAX)

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

func testCreateExternalIncentiveGns() {
	std.TestSetRealm(adminRealm)
	gns.Transfer(externalCreatorUser, 9000000000)
	gns.Transfer(externalCreatorUser, depositGnsAmount)

	std.TestSetRealm(externalCreatorRealm) // creator
	gns.Approve(stakerAddr, consts.UINT64_MAX)

	std.TestSkipHeights(1)
	sr.CreateExternalIncentive(
		"gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100",
		gnsPath,
		9000000000,
		1234569600,
		1234569600+TIMESTAMP_90DAYS,
	)
}

func testMakeExternalStart() {
	externalStartTime := int64(1234569600)
	nowTime := time.Now().Unix()
	timeLeft := externalStartTime - nowTime

	blockAvgTime := milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockLeft := timeLeft / blockAvgTime

	std.TestSkipHeights(int64(blockLeft)) // skip until external bar starts
	std.TestSkipHeights(10)               // skip bit more to see reward calculation
}

func testStakeTokenPos01() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(1)
	gnft.Approve(stakerAddr, positionIdFrom(1))
	sr.StakeToken(1, "")
}

func testStakeTokenPos02() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(1)
	gnft.Approve(stakerAddr, positionIdFrom(2))
	sr.StakeToken(2, "")
}

func testCollectRewardPos01AndPos02() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(1)
	sr.CollectReward(1, false)
	sr.CollectReward(2, false)
}

func testEndExternalGns() {
	externalEndTime := (1234569600 + TIMESTAMP_90DAYS)
	nowTime := time.Now().Unix()
	timeLeft := externalEndTime - nowTime

	blockAvgTime := milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockLeft := timeLeft / blockAvgTime

	std.TestSkipHeights(int64(blockLeft)) // skip until external gns ends
	std.TestSkipHeights(10)               // skip bit more to see reward calculation

	std.TestSetRealm(externalCreatorRealm)
	gnsBalanceBeforeEnds := gns.BalanceOf(externalCreatorUser)
	sr.EndExternalIncentive(
		externalCreatorAddr,
		"gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100",
		gnsPath,
		1234569600,
		1234569600+TIMESTAMP_90DAYS,
		126,
	)
	gnsBalanceAfterEnds := gns.BalanceOf(externalCreatorUser)
	if gnsBalanceAfterEnds != 1000028693 {
		// 1000000000 = depositGnsAmount
		// 28693 = penalty
		panic("expected 1000028693")
	}
}

func testCollectRewardPos01AndPos02AfterEnd() {
	std.TestSetRealm(adminRealm)
	std.TestSkipHeights(1)

	before := gns.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	sr.CollectReward(2, false)
	after := gns.BalanceOf(adminUser)
	diff := after - before
	if diff == 0 {
		panic("reward can not be zero")
	}
}

func positionIdFrom(positionId interface{}) grc721.TokenID {
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
