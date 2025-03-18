// PKGPATH: gno.land/r/gnoswap/v1/staker_test

// POOLs:
// 1. gnot:gns:3000

// POSITIONs:
// 1. in-range
// 2. in-range

// REWARDs:
// - external bar ( bar:qux:100 )
// - external qux ( qux:gns:100 )

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
	adminRealm   = std.NewUserRealm(adminAddr)

	stakerAddr, _ = access.GetAddress(access.ROLE_STAKER)
	poolAddr, _   = access.GetAddress(access.ROLE_POOL)
	wugnotAddr    = consts.WUGNOT_ADDR

	barPath = "gno.land/r/onbloc/bar"
	quxPath = "gno.land/r/onbloc/qux"

	fee100      uint32 = 100
	max_timeout int64  = 9999999999

	// external incentive deposit fee
	depositGnsAmount uint64 = 1_000_000_000 // 1_000 GNS

	TIMESTAMP_90DAYS int64 = 90 * 24 * 60 * 60
)

func main() {
	testInit()
	testCreatePool()

	testCreateExternalIncentiveBar()
	testCreateExternalIncentiveQux()

	testMintAndStakeBarQuxPos01()
	testMintAndStakeBarQuxPos02()

	testMakeExternalBarAndQuxStart()

	testCollectRewardPos01()
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
		barPath,
		quxPath,
		fee100,
		common.TickMathGetSqrtRatioAtTick(0).ToString(), // 79228162514264337593543950337
	)
}

func testCreateExternalIncentiveBar() {
	std.TestSetRealm(adminRealm)

	bar.Approve(stakerAddr, consts.UINT64_MAX)
	gns.Approve(stakerAddr, depositGnsAmount)

	std.TestSkipHeights(1)
	sr.CreateExternalIncentive(
		"gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100",
		barPath,
		9000000000,
		1234569600,
		1234569600+TIMESTAMP_90DAYS,
	)
}

func testCreateExternalIncentiveQux() {
	std.TestSetRealm(adminRealm)

	qux.Approve(stakerAddr, consts.UINT64_MAX)
	gns.Approve(stakerAddr, depositGnsAmount)

	std.TestSkipHeights(1)
	sr.CreateExternalIncentive(
		"gno.land/r/onbloc/bar:gno.land/r/onbloc/qux:100",
		quxPath,
		18000000000,
		1234569600,
		1234569600+TIMESTAMP_90DAYS,
	)
}

func testMintAndStakeBarQuxPos01() {
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
}

func testMintAndStakeBarQuxPos02() {
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
}

func testMakeExternalBarAndQuxStart() {
	externalStartTime := int64(1234569600)
	nowTime := time.Now().Unix()
	timeLeft := externalStartTime - nowTime

	blockAvgTime := milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockLeft := timeLeft / blockAvgTime

	std.TestSkipHeights(int64(blockLeft)) // skip until external bar starts
	std.TestSkipHeights(10)               // skip bit more to see reward calculation

	if extractReward(1) != extractReward(2) {
		panic("reward is not equal")
	}
}

func testCollectRewardPos01() {
	std.TestSetRealm(adminRealm)

	std.TestSkipHeights(10)
	sr.CollectReward(1, false)

	reward01 := extractReward(1)
	if reward01 != 0 {
		panic("reward collected, should be 0 for checking")
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

func extractReward(positionId uint64) uint64 {
	apiReward := sr.ApiGetRewardsByLpTokenId(positionId)
	rawReward, _ := json.Unmarshal([]byte(apiReward))
	rawRewardObject, _ := rawReward.GetKey("response")
	arrReward, _ := rawRewardObject.GetArray()

	reward, _ := arrReward[0].GetKey("rewards")
	rewardArr, _ := reward.GetArray()
	if len(rewardArr) == 0 {
		return 0
	}
	rewardTokenAmount, _ := rewardArr[0].GetKey("rewardTokenAmount")

	rewardTokenAmountInt, _ := strconv.ParseUint(rewardTokenAmount.String(), 10, 64)
	return rewardTokenAmountInt
}

func milliToSec(ms int64) int64 {
	return ms / consts.MILLISECONDS_PER_SECOND
}
