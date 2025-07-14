// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. bar:qux:100

// POSITIONs:
// 1. in-range -> out-range -> in-range
// 2. (always) in-range

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

	stakerAddr, _ = access.GetAddress(access.ROLE_STAKER)
	poolAddr, _   = access.GetAddress(access.ROLE_POOL)
	routerAddr, _ = access.GetAddress(access.ROLE_ROUTER)

	barPath = "gno.land/r/onbloc/bar"
	quxPath = "gno.land/r/onbloc/qux"

	fee100      uint32 = 100
	max_timeout int64  = 9999999999

	// external incentive deposit fee
	depositGnsAmount uint64 = 1_000_000_000 // 1_000 GNS

	TIMESTAMP_90DAYS int64 = 90 * 24 * 60 * 60
	poolPath               = "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000"
)

func main() {
	testInit()
	testCreatePool()
	testMintBarQuxPos01()
	testMintBarQuxPos02()
	testCreateExternalIncentive()
	testStakeTokenPos01AndPos02()
	testMakeExternalBarStart()

	testCheckReward01() // both positions are in-range

	testMakePosition1OutRangeBySwap()
	testCheckReward02() // position-01 is out-range

	testMakePosition1InRangeBySwap() // position-01 is in-range again
	testCheckReward03()              // both positions are in-range
}

func testInit() {
	testing.SetRealm(adminRealm)
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
		int32(-1000),
		int32(1000),
		"500000",
		"500000",
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

func testStakeTokenPos01AndPos02() {
	testing.SetRealm(adminRealm)

	gnft.Approve(stakerAddr, positionIdFrom(1))
	gnft.Approve(stakerAddr, positionIdFrom(2))

	testing.SkipHeights(1)
	sr.StakeToken(1, "")
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
}

func testCheckReward01() {
	testing.SetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar

	if diff == 0 {
		panic("position 01 is in-range, should have reward")
	}
}

func testMakePosition1OutRangeBySwap() {
	testing.SetRealm(adminRealm)

	poolTick := pl.PoolGetSlot0Tick(poolPath)

	bar.Approve(poolAddr, consts.UINT64_MAX)
	qux.Approve(poolAddr, consts.UINT64_MAX)

	bar.Approve(routerAddr, consts.UINT64_MAX)
	qux.Approve(routerAddr, consts.UINT64_MAX)

	tokenIn, tokenOut := cross(rr.ExactInSwapRoute)(
		barPath,     // inputToken
		quxPath,     // outputToken
		"100000",    // finalAmountIn
		poolPath,    // RouteArr
		"100",       // quoteArr
		"0",         // amountOutMin
		max_timeout, // timeout
		"",          // referrer
	)
	testing.SkipHeights(1)

	newPoolTick := pl.PoolGetSlot0Tick(poolPath)
	println("oldPoolTick", poolTick)
	println("newPoolTick", newPoolTick)
	println()
}

func testCheckReward02() {
	testing.SetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar

	if diff != 0 {
		panic("position 01 is out-range, should not have reward")
	}
}

func testMakePosition1InRangeBySwap() {
	testing.SetRealm(adminRealm)

	poolTick := pl.PoolGetSlot0Tick(poolPath)

	bar.Approve(poolAddr, consts.UINT64_MAX)
	qux.Approve(poolAddr, consts.UINT64_MAX)

	bar.Approve(routerAddr, consts.UINT64_MAX)
	qux.Approve(routerAddr, consts.UINT64_MAX)

	tokenIn, tokenOut := cross(rr.ExactInSwapRoute)(
		quxPath,
		barPath,
		"100000",
		"gno.land/r/onbloc/qux:gno.land/r/onbloc/bar:100",
		"100",
		"0",
		max_timeout,
		"",
	)

	newPoolTick := pl.PoolGetSlot0Tick(poolPath)
	println("oldPoolTick", poolTick)
	println("newPoolTick", newPoolTick)
	println()
}

func testCheckReward03() {
	testing.SkipHeights(1)

	testing.SetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar

	if diff <= 0 {
		panic("position 01 is in-range, should have reward")
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
