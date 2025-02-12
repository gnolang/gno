// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. gnot:gns:3000

// POSITIONs:
// 1. in-range -> out-range -> in-range
// 2. (always) in-range

// REWARDs:
// - internal tier 1 ( gnot:gns:3000 )

package staker_test

import (
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gns"

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

	wugnotAddr = consts.WUGNOT_ADDR

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

	poolPath = "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000"
)

func main() {
	testInit()
	testCreatePool()
	testMintWugnotGnsPos01()
	testMintWugnotGnsPos02()

	testStakeTokenPos01AndPos02()

	testCheckReward01() // both positions are in-range

	testMakePosition1OutRangeBySwap()
	testCheckReward02() // position-01 is out-range

	testMakePosition1InRangeBySwap() // position-01 is in-range again
	testCheckReward03()              // both positions are in-range
}

func testInit() {
	std.TestSetRealm(adminRealm)

	// prepare wugnot
	std.TestIssueCoins(adminAddr, std.Coins{{"ugnot", 100_000_000_000_000}})
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(adminAddr, wugnotAddr, std.Coins{{"ugnot", 50_000_000_000_000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 50_000_000_000_000}}, nil)
	wugnot.Deposit()
	std.TestSetOrigSend(nil, nil)
}

func testCreatePool() {
	std.TestSetRealm(adminRealm)

	pl.SetPoolCreationFeeByAdmin(0)

	std.TestSkipHeights(1)
	pl.CreatePool(
		wugnotPath,
		gnsPath,
		fee3000,
		common.TickMathGetSqrtRatioAtTick(0).ToString(), // 79228162514264337593543950337
	)
}

func testMintWugnotGnsPos01() {
	std.TestSetRealm(adminRealm)

	wugnot.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	std.TestSkipHeights(1)
	pn.Mint(
		wugnotPath,
		gnsPath,
		fee3000,
		int32(-60),
		int32(60),
		"50",
		"50",
		"1",
		"1",
		max_timeout,
		adminAddr,
		adminAddr,
	)
}

func testMintWugnotGnsPos02() {
	std.TestSetRealm(adminRealm)

	wugnot.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	std.TestSkipHeights(1)
	pn.Mint(
		wugnotPath,
		gnsPath,
		fee3000,
		int32(-1020),
		int32(1020),
		"500000",
		"500000",
		"1",
		"1",
		max_timeout,
		adminAddr,
		adminAddr,
	)
}

func testStakeTokenPos01AndPos02() {
	std.TestSetRealm(adminRealm)

	gnft.Approve(stakerAddr, tokenIdFrom(1))
	gnft.Approve(stakerAddr, tokenIdFrom(2))

	std.TestSkipHeights(1)
	sr.StakeToken(1)
	sr.StakeToken(2)
}

func testCheckReward01() {
	std.TestSkipHeights(1)
	std.TestSetRealm(adminRealm)

	oldGns := gns.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newGns := gns.BalanceOf(adminUser)
	diff := newGns - oldGns

	if diff == 0 {
		panic("position 01 is in-range, should have reward")
	}
}

func testMakePosition1OutRangeBySwap() {
	std.TestSetRealm(adminRealm)

	poolTick := pl.PoolGetSlot0Tick(poolPath)

	wugnot.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	wugnot.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)

	tokenIn, tokenOut := rr.ExactInSwapRoute(
		wugnotPath, // inputToken
		gnsPath,    // outputToken
		"100000",   // finalAmountIn
		poolPath,   // RouteArr
		"100",      // quoteArr
		"0",        // amountOutMin
		max_timeout,
	)

	newPoolTick := pl.PoolGetSlot0Tick(poolPath)
	println("oldPoolTick", poolTick)
	println("newPoolTick", newPoolTick)
	println()
}

func testCheckReward02() {
	std.TestSkipHeights(1)
	std.TestSetRealm(adminRealm)

	oldGns := gns.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newGns := gns.BalanceOf(adminUser)
	diff := newGns - oldGns

	if diff != 0 {
		panic("position 01 is out-range, should not have reward")
	}
}

func testMakePosition1InRangeBySwap() {
	std.TestSetRealm(adminRealm)

	poolTick := pl.PoolGetSlot0Tick(poolPath)

	wugnot.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	wugnot.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)
	gns.Approve(consts.ROUTER_ADDR, consts.UINT64_MAX)

	tokenIn, tokenOut := rr.ExactInSwapRoute(
		gnsPath,
		wugnotPath,
		"100000",
		"gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:3000",
		"100",
		"0",
		max_timeout,
	)

	newPoolTick := pl.PoolGetSlot0Tick(poolPath)
	println("oldPoolTick", poolTick)
	println("newPoolTick", newPoolTick)
	println()
}

func testCheckReward03() {
	std.TestSkipHeights(1)
	std.TestSetRealm(adminRealm)

	oldGns := gns.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	newGns := gns.BalanceOf(adminUser)
	diff := newGns - oldGns

	if diff == 0 {
		panic("position 01 is in-range, should have reward")
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
