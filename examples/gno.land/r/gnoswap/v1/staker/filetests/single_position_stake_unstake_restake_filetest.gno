// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. gnot:gns:3000

// POSITIONs:
// 1. in-range ( stake -> unstake -> restake)

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

	wugnotAddr = consts.WUGNOT_ADDR

	gnsPath    = "gno.land/r/gnoswap/v1/gns"
	wugnotPath = "gno.land/r/demo/wugnot"

	fee3000     uint32 = 3000
	max_timeout int64  = 9999999999
)

func main() {
	testInit()
	testCreatePool()
	testMintWugnotGnsPos01()
	testMintWugnotGnsPos02()

	testStakeTokenPos01AndPos02()

	testUnstakeTokenPos01()

	testStakeTokenPos01Again()
}

func testInit() {
	std.TestSetRealm(adminRealm)

	// prepare wugnot
	std.TestIssueCoins(adminAddr, std.Coins{{"ugnot", 100_000_000_000_000}})
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(adminAddr, wugnotAddr, std.Coins{{"ugnot", 50_000_000_000_000}})
	std.TestSetOriginSend(std.Coins{{"ugnot", 50_000_000_000_000}}, nil)
	wugnot.Deposit()
	std.TestSetOriginSend(nil, nil)
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

	wugnot.Approve(poolAddr, consts.UINT64_MAX)
	gns.Approve(poolAddr, consts.UINT64_MAX)

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

	wugnot.Approve(poolAddr, consts.UINT64_MAX)
	gns.Approve(poolAddr, consts.UINT64_MAX)

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

	gnft.Approve(stakerAddr, positionIdFrom(1))
	gnft.Approve(stakerAddr, positionIdFrom(2))

	std.TestSkipHeights(1)
	sr.StakeToken(1, "")
	sr.StakeToken(2, "")
}

func testUnstakeTokenPos01() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(1)

	beforeGns := gns.BalanceOf(adminUser)
	sr.UnStakeToken(1, false)
	afterGns := gns.BalanceOf(adminUser)
	diff := afterGns - beforeGns

	if diff == 0 {
		panic("position 01 was in-range, should have reward")
	}
}

func testStakeTokenPos01Again() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(1)
	gnft.Approve(stakerAddr, positionIdFrom(1))
	sr.StakeToken(1, "")

	// reward check for staked > unstaked > staked position
	std.TestSkipHeights(1)
	beforeGns := gns.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	afterGns := gns.BalanceOf(adminUser)
	diff := afterGns - beforeGns

	if diff == 0 {
		panic("position 01 in-range, should have reward")
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
