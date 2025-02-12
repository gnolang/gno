// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. gnot:gns:3000
// 2. bar:baz:100

// POSITIONs:
// 1. in-range ( gnot:gns:3000 )
// 2. in-range ( bar:baz:100 )

// REWARDs:
// - internal tier 1 ( gnot:gns:3000 )
// - internal tier 2 ( bar:baz:100 ) -> will be removed from internal tiers

package staker_test

import (
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"

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
)

func main() {
	testInit()
	testCreatePool()

	testMintAndStakeWugnotGnsPos01()

	testSetPoolTier2() // new pool is set to tier 2
	testMintAndStakeBarBazPos02()

	testRemovePoolTier()
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

	// set unstaking fee to 0
	sr.SetUnStakingFeeByAdmin(0)
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
	pl.CreatePool(
		barPath,
		bazPath,
		fee100,
		common.TickMathGetSqrtRatioAtTick(0).ToString(), // 79228162514264337593543950337
	)
}

func testMintAndStakeWugnotGnsPos01() {
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
		"100",
		"100",
		"0",
		"0",
		max_timeout,
		adminAddr,
		adminAddr,
	)

	gnft.Approve(stakerAddr, positionIdFrom(1))
	sr.StakeToken(1, "")

	// check reward for position 01 (gnot:gns:3000 tier 01)
	std.TestSetRealm(adminRealm)
	std.TestSkipHeights(1)
	beforeGns := gns.BalanceOf(adminUser)
	sr.CollectReward(1, false)
	afterGns := gns.BalanceOf(adminUser)
	diff := afterGns - beforeGns
	if !isInErrorRange(3210615, diff) {
		panic("expected about 3210615") // 10702054 * 100%(tier1 ratio) * 30%(warmUp)
	}
}

func testSetPoolTier2() {
	std.TestSetRealm(adminRealm)

	sr.SetPoolTierByAdmin("gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:100", 2)
}

func testMintAndStakeBarBazPos02() {
	std.TestSetRealm(adminRealm)

	bar.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
	baz.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

	pn.Mint(
		barPath,
		bazPath,
		fee100,
		int32(-50),
		int32(50),
		"100",
		"100",
		"0",
		"0",
		max_timeout,
		adminAddr,
		adminAddr,
	)

	gnft.Approve(stakerAddr, positionIdFrom(2))
	sr.StakeToken(2, "")
	std.TestSkipHeights(1)

	{
		// check reward for position 01 (gnot:gns:3000 tier 01)
		std.TestSetRealm(adminRealm)
		beforeGns := gns.BalanceOf(adminUser)
		sr.CollectReward(1, false)
		afterGns := gns.BalanceOf(adminUser)
		diff := afterGns - beforeGns
		if !isInErrorRange(2247430, diff) {
			panic("expected about 2247430") // 10702054 * 70%(tier1 ratio) * 30%(warmUp)
		}
	}

	{
		// check reward for position 02 (bar:baz:100 tier 02)
		std.TestSetRealm(adminRealm)
		beforeGns := gns.BalanceOf(adminUser)
		sr.CollectReward(2, false)
		afterGns := gns.BalanceOf(adminUser)
		diff := afterGns - beforeGns
		if !isInErrorRange(963184, diff) {
			panic("expected about 963184") // 10702054 * 30%(tier2 ratio) * 30% (warmUp)
		}
	}
}

func testRemovePoolTier() {
	std.TestSetRealm(adminRealm)

	sr.RemovePoolTierByAdmin("gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:100")
	std.TestSkipHeights(1)

	{
		// check reward for position 01 (gnot:gns:3000 tier 01)
		std.TestSetRealm(adminRealm)
		beforeGns := gns.BalanceOf(adminUser)
		sr.CollectReward(1, false)
		afterGns := gns.BalanceOf(adminUser)
		diff := afterGns - beforeGns
		if !isInErrorRange(3210615, diff) {
			panic("expected about 3210615") // 10702054 * 100%(tier1 ratio) * 30%(warmUp)
		}
	}

	{
		// check reward for position 02 (bar:baz:100 tier 02)
		std.TestSetRealm(adminRealm)
		beforeGns := gns.BalanceOf(adminUser)
		sr.CollectReward(2, false)
		afterGns := gns.BalanceOf(adminUser)
		diff := afterGns - beforeGns
		if diff != 0 {
			panic("expected about 0") // bar:baz:100 is removed from internal tiers
		}
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

func isInErrorRange(expected uint64, actual uint64) bool {
	lowerBound := expected * 999999 / 1000000
	upperBound := expected * 1000001 / 1000000
	return actual >= lowerBound && actual <= upperBound
}
