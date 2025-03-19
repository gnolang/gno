// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. gnot:gns:3000

// POSITIONs:
// 1. in-range ( will be unstaked )

// REWARDs:
// - internal tier 1 ( gnot:gns:3000 )
// - halving changes

package staker_test

import (
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/testutils"

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

	// g1v4u8getjdeskcsmjv4shgmmjta047h6lua7mup
	externalCreatorAddr  = testutils.TestAddress("externalCreator")
	externalCreatorUser  = externalCreatorAddr
	externalCreatorRealm = std.NewUserRealm(externalCreatorAddr)

	stakerAddr, _ = access.GetAddress(access.ROLE_STAKER)
	stakerUser    = stakerAddr
	stakerRealm   = std.NewCodeRealm(consts.STAKER_PATH)

	poolAddr, _ = access.GetAddress(access.ROLE_POOL)

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
	testStakeTokenPos01()
	testGnsAvgBlockTime()
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

func testStakeTokenPos01() {
	std.TestSetRealm(adminRealm)

	// stake
	gnft.Approve(stakerAddr, positionIdFrom(1))
	sr.StakeToken(1, "")

	before := gns.BalanceOf(adminUser)

	std.TestSkipHeights(100)
	sr.CollectReward(1, false)

	after := gns.BalanceOf(adminUser)

	diff := after - before
	if diff <= 0 {
		panic("reward can not be 0 for 100 blocks")
	}
}

func testGnsAvgBlockTime() {
	std.TestSetRealm(adminRealm)

	before := gns.BalanceOf(adminUser)
	std.TestSkipHeights(1)
	sr.CollectReward(1, false)
	after := gns.BalanceOf(adminUser)

	diff := after - before // actual reward for 1 block when avg block time was 2000ms(2s)
	if !isInErrorRange(3210616, diff) {
		panic("expected about 3210616")
	}

	gns.SetAvgBlockTimeInMsByAdmin(4000) // orig block time was 2000ms(2s), change it to 4s
	sr.CollectReward(1, false)           // clear reward

	before = gns.BalanceOf(adminUser)
	std.TestSkipHeights(1)
	sr.CollectReward(1, false) // actual reward for 1 block when avg block time was 4000ms(4s)
	after = gns.BalanceOf(adminUser)
	diff = after - before
	if !isInErrorRange(6421232, diff) {
		panic("expected about 6421232") // block time has doubled, so reward should be doubled too
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
