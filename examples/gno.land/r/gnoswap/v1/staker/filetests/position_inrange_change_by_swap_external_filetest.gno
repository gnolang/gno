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
	rr "gno.land/r/gnoswap/v1/router"
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
	testStakeTokenPos01AndPos02()
	testMakeExternalBarStart()

	testCheckReward01() // both positions are in-range

	testMakePosition1OutRangeBySwap()
	testCheckReward02() // position-01 is out-range

	testMakePosition1InRangeBySwap() // position-01 is in-range again
	testCheckReward03()              // both positions are in-range
}

func testInit() {
	std.TestSetRealm(adminRealm)
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

func testStakeTokenPos01AndPos02() {
	std.TestSetRealm(adminRealm)

	gnft.Approve(stakerAddr, positionIdFrom(1))
	gnft.Approve(stakerAddr, positionIdFrom(2))

	std.TestSkipHeights(1)
	sr.StakeToken(1, "")
	sr.StakeToken(2, "")
}

func testMakeExternalBarStart() {
	externalStartTime := int64(1234569600)
	nowTime := time.Now().Unix()
	timeLeft := externalStartTime - nowTime

	blockAvgTime := milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockLeft := timeLeft / blockAvgTime

	std.TestSkipHeights(int64(blockLeft)) // skip until external bar starts
	std.TestSkipHeights(10)               // skip bit more to see reward calculation
}

func testCheckReward01() {
	std.TestSetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar

	if diff == 0 {
		panic("position 01 is in-range, should have reward")
	}
}

func testMakePosition1OutRangeBySwap() {
	std.TestSetRealm(adminRealm)

	poolTick := pl.PoolGetSlot0Tick(poolPath)

	bar.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	qux.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	bar.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)
	qux.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)

	tokenIn, tokenOut := rr.ExactInSwapRoute(
		barPath,     // inputToken
		quxPath,     // outputToken
		"100000",    // finalAmountIn
		poolPath,    // RouteArr
		"100",       // quoteArr
		"0",         // amountOutMin
		max_timeout, // timeout
		"",          // referrer
	)
	std.TestSkipHeights(1)

	newPoolTick := pl.PoolGetSlot0Tick(poolPath)
	println("oldPoolTick", poolTick)
	println("newPoolTick", newPoolTick)
	println()
}

func testCheckReward02() {
	std.TestSetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar

	if diff != 0 {
		panic("position 01 is out-range, should not have reward")
	}
}

func testMakePosition1InRangeBySwap() {
	std.TestSetRealm(adminRealm)

	poolTick := pl.PoolGetSlot0Tick(poolPath)

	bar.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	qux.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	bar.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)
	qux.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)

	tokenIn, tokenOut := rr.ExactInSwapRoute(
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
	std.TestSkipHeights(1)

	std.TestSetRealm(adminRealm)

	oldBar := bar.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newBar := bar.BalanceOf(adminUser)
	diff := newBar - oldBar

	if diff <= 0 {
		panic("position 01 is in-range, should have reward")
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
