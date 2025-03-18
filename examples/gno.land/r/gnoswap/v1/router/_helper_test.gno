package router

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/gnoswap/consts"
	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"
	sr "gno.land/r/gnoswap/v1/staker"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	"gno.land/r/onbloc/foo"
	"gno.land/r/onbloc/obl"
	"gno.land/r/onbloc/qux"

	"gno.land/r/gnoswap/v1/access"
)

const (
	ugnotDenom string = "ugnot"
	ugnotPath  string = "ugnot"
	wugnotPath string = "gno.land/r/demo/wugnot"
	gnsPath    string = "gno.land/r/gnoswap/v1/gns"
	barPath    string = "gno.land/r/onbloc/bar"
	bazPath    string = "gno.land/r/onbloc/baz"
	fooPath    string = "gno.land/r/onbloc/foo"
	oblPath    string = "gno.land/r/onbloc/obl"
	quxPath    string = "gno.land/r/onbloc/qux"

	fee100      uint32 = 100
	fee500      uint32 = 500
	fee3000     uint32 = 3000
	maxApprove  uint64 = 18446744073709551615
	max_timeout int64  = 9999999999

	TIER_1 uint64 = 1
	TIER_2 uint64 = 2
	TIER_3 uint64 = 3

	poolCreationFee = 100_000_000
)

const (
	FEE_LOW    uint32 = 500
	FEE_MEDIUM uint32 = 3000
	FEE_HIGH   uint32 = 10000
)

var (
	// define addresses to use in tests
	addr01 = testutils.TestAddress("addr01")
	addr02 = testutils.TestAddress("addr02")
)

var (
	user1Addr       std.Address = "g1ecely4gjy0yl6s9kt409ll330q9hk2lj9ls3ec"
	singlePoolPath              = "gno.land/r/onbloc/bar:gno.land/r/onbloc/baz:3000"
	singlePoolPath2             = "gno.land/r/onbloc/baz:gno.land/r/onbloc/bar:3000"
)

var (
	minTick int32 = -887220
	maxTick int32 = 887220
)

var (
	admin, _  = access.GetAddress(access.ROLE_ADMIN)
	adminAddr = admin

	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")

	adminRealm  = std.NewUserRealm(admin)
	routerRealm = std.NewUserRealm(routerAddr)
	posRealm    = std.NewCodeRealm(consts.POSITION_PATH)

	// addresses used in tests
	addrUsedInTest = []std.Address{addr01, addr02}
)

func InitialisePoolTest(t *testing.T) {
	t.Helper()

	ugnotFaucet(t, admin, 100_000_000_000_000)
	ugnotDeposit(t, admin, 100_000_000_000_000)

	std.TestSetOriginCaller(admin)
	TokenApprove(t, gnsPath, admin, poolAddr, maxApprove)
	CreatePool(t, wugnotPath, gnsPath, fee3000, "79228162514264337593543950336", admin)

	// 2. create position
	std.TestSetOriginCaller(alice)
	TokenFaucet(t, wugnotPath, alice)
	TokenFaucet(t, gnsPath, alice)
	TokenApprove(t, wugnotPath, alice, poolAddr, uint64(1000))
	TokenApprove(t, gnsPath, alice, poolAddr, uint64(1000))
}

func TokenFaucet(t *testing.T, tokenPath string, to std.Address) {
	t.Helper()
	std.TestSetOriginCaller(admin)
	defaultAmount := uint64(5_000_000_000)

	switch tokenPath {
	case wugnotPath:
		wugnotTransfer(t, to, defaultAmount)
	case gnsPath:
		gnsTransfer(t, to, defaultAmount)
	case barPath:
		barTransfer(t, to, defaultAmount)
	case bazPath:
		bazTransfer(t, to, defaultAmount)
	case fooPath:
		fooTransfer(t, to, defaultAmount)
	case oblPath:
		oblTransfer(t, to, defaultAmount)
	case quxPath:
		quxTransfer(t, to, defaultAmount)
	default:
		panic("token not found")
	}
}

func TokenBalance(t *testing.T, tokenPath string, owner std.Address) uint64 {
	t.Helper()
	switch tokenPath {
	case wugnotPath:
		return wugnot.BalanceOf(owner)
	case gnsPath:
		return gns.BalanceOf(owner)
	case barPath:
		return bar.BalanceOf(owner)
	case bazPath:
		return baz.BalanceOf(owner)
	case fooPath:
		return foo.BalanceOf(owner)
	case oblPath:
		return obl.BalanceOf(owner)
	case quxPath:
		return qux.BalanceOf(owner)
	default:
		panic("token not found")
	}
}

func TokenAllowance(t *testing.T, tokenPath string, owner, spender std.Address) uint64 {
	t.Helper()
	switch tokenPath {
	case wugnotPath:
		return wugnot.Allowance(owner, spender)
	case gnsPath:
		return gns.Allowance(owner, spender)
	case barPath:
		return bar.Allowance(owner, spender)
	case bazPath:
		return baz.Allowance(owner, spender)
	case fooPath:
		return foo.Allowance(owner, spender)
	case oblPath:
		return obl.Allowance(owner, spender)
	case quxPath:
		return qux.Allowance(owner, spender)
	default:
		panic("token not found")
	}
}

func TokenApprove(t *testing.T, tokenPath string, owner, spender std.Address, amount uint64) {
	t.Helper()
	switch tokenPath {
	case wugnotPath:
		wugnotApprove(t, owner, spender, amount)
	case gnsPath:
		gnsApprove(t, owner, spender, amount)
	case barPath:
		barApprove(t, owner, spender, amount)
	case bazPath:
		bazApprove(t, owner, spender, amount)
	case fooPath:
		fooApprove(t, owner, spender, amount)
	case oblPath:
		oblApprove(t, owner, spender, amount)
	case quxPath:
		quxApprove(t, owner, spender, amount)
	default:
		panic("token not found")
	}
}

func CreatePool(t *testing.T,
	token0 string,
	token1 string,
	fee uint32,
	sqrtPriceX96 string,
	caller std.Address,
) {
	t.Helper()

	std.TestSetRealm(std.NewUserRealm(caller))
	poolPath := pl.GetPoolPath(token0, token1, fee)
	if !pl.DoesPoolPathExist(poolPath) {
		pl.CreatePool(token0, token1, fee, sqrtPriceX96)
		sr.SetPoolTierByAdmin(poolPath, TIER_1)
	}
}

func LPTokenStake(t *testing.T, owner std.Address, positionId uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	sr.StakeToken(positionId, "")
}

func LPTokenUnStake(t *testing.T, owner std.Address, positionId uint64, unwrap bool) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	sr.UnStakeToken(positionId, unwrap)
}

func wugnotApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	wugnot.Approve(spender, amount)
}

func gnsApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	gns.Approve(spender, amount)
}

func barApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	bar.Approve(spender, amount)
}

func bazApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	baz.Approve(spender, amount)
}

func fooApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	foo.Approve(spender, amount)
}

func oblApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	obl.Approve(spender, amount)
}

func quxApprove(t *testing.T, owner, spender std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(owner))
	qux.Approve(spender, amount)
}

func wugnotTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	wugnot.Transfer(to, amount)
}

func gnsTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	gns.Transfer(to, amount)
}

func barTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	bar.Transfer(to, amount)
}

func bazTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	baz.Transfer(to, amount)
}

func fooTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	foo.Transfer(to, amount)
}

func oblTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	obl.Transfer(to, amount)
}

func quxTransfer(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(admin))
	qux.Transfer(to, amount)
}

// ----------------------------------------------------------------------------
// ugnot

func ugnotTransfer(t *testing.T, from, to std.Address, amount uint64) {
	t.Helper()

	std.TestSetRealm(std.NewUserRealm(from))
	std.TestSetOriginSend(std.Coins{{ugnotDenom, int64(amount)}}, nil)
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(from, to, std.Coins{{ugnotDenom, int64(amount)}})
}

func ugnotBalanceOf(t *testing.T, addr std.Address) uint64 {
	t.Helper()

	banker := std.NewBanker(std.BankerTypeRealmIssue)
	coins := banker.GetCoins(addr)
	if len(coins) == 0 {
		return 0
	}

	return uint64(coins.AmountOf(ugnotDenom))
}

func ugnotMint(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	banker := std.NewBanker(std.BankerTypeRealmIssue)
	banker.IssueCoin(addr, denom, amount)
	std.TestIssueCoins(addr, std.Coins{{denom, int64(amount)}})
}

func ugnotBurn(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	banker := std.NewBanker(std.BankerTypeRealmIssue)
	banker.RemoveCoin(addr, denom, amount)
}

func ugnotFaucet(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	faucetAddress := admin
	std.TestSetOriginCaller(faucetAddress)

	if ugnotBalanceOf(t, faucetAddress) < amount {
		newCoins := std.Coins{{ugnotDenom, int64(amount)}}
		ugnotMint(t, faucetAddress, newCoins[0].Denom, newCoins[0].Amount)
		std.TestSetOriginSend(newCoins, nil)
	}
	ugnotTransfer(t, faucetAddress, to, amount)
}

func ugnotDeposit(t *testing.T, addr std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(addr))
	wugnotAddr := consts.WUGNOT_ADDR
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(addr, wugnotAddr, std.Coins{{ugnotDenom, int64(amount)}})
	wugnot.Deposit()
}

func CreatePoolWithoutFee(t *testing.T) {
	std.TestSetRealm(adminRealm)
	// set pool create fee to 0 for testing
	pl.SetPoolCreationFeeByAdmin(0)
	CreatePool(t, barPath, fooPath, fee500, common.TickMathGetSqrtRatioAtTick(0).ToString(), admin)
	CreatePool(t, bazPath, fooPath, fee3000, common.TickMathGetSqrtRatioAtTick(0).ToString(), admin)
	CreatePool(t, barPath, bazPath, fee3000, common.TickMathGetSqrtRatioAtTick(0).ToString(), admin)
	CreatePool(t, barPath, bazPath, fee500, common.TickMathGetSqrtRatioAtTick(0).ToString(), admin)
}

func CreateSecondPoolWithoutFee(t *testing.T) {
	std.TestSetRealm(adminRealm)
	pl.SetPoolCreationFeeByAdmin(0)

	CreatePool(t,
		bazPath,
		quxPath,
		fee3000,
		common.TickMathGetSqrtRatioAtTick(0).ToString(),
		admin,
	)
}

func MakeMintPositionWithoutFee(t *testing.T) (uint64, string, string, string) {
	t.Helper()

	// make actual data to test resetting not only position's state but also pool's state
	std.TestSetRealm(adminRealm)

	TokenApprove(t, barPath, admin, poolAddr, consts.UINT64_MAX)
	TokenApprove(t, bazPath, admin, poolAddr, consts.UINT64_MAX)

	// mint position
	return pn.Mint(
		barPath,
		bazPath,
		fee3000,
		-887220,
		887220,
		"50000",
		"50000",
		"0",
		"0",
		max_timeout,
		admin,
		admin,
		"",
	)
}

func MakeSecondMintPositionWithoutFee(t *testing.T) (uint64, string, string, string) {
	t.Helper()

	std.TestSetRealm(adminRealm)

	TokenApprove(t, bazPath, admin, poolAddr, consts.UINT64_MAX)
	TokenApprove(t, quxPath, admin, poolAddr, consts.UINT64_MAX)

	return pn.Mint(
		bazPath,
		quxPath,
		fee3000,
		-887220,
		887220,
		"50000",
		"50000",
		"0",
		"0",
		max_timeout,
		admin,
		admin,
		"",
	)
}

func MakeThirdMintPositionWithoutFee(t *testing.T) (uint64, string, string, string) {
	t.Helper()

	std.TestSetRealm(adminRealm)

	TokenApprove(t, barPath, admin, poolAddr, consts.UINT64_MAX)
	TokenApprove(t, fooPath, admin, poolAddr, consts.UINT64_MAX)

	return pn.Mint(
		barPath,
		fooPath,
		fee500,
		-887220,
		887220,
		"50000",
		"50000",
		"0",
		"0",
		max_timeout,
		admin,
		admin,
		"",
	)
}

func MakeForthMintPositionWithoutFee(t *testing.T) (uint64, string, string, string) {
	t.Helper()

	// make actual data to test resetting not only position's state but also pool's state
	std.TestSetRealm(adminRealm)

	TokenApprove(t, barPath, admin, poolAddr, consts.UINT64_MAX)
	TokenApprove(t, bazPath, admin, poolAddr, consts.UINT64_MAX)

	// mint position
	return pn.Mint(
		barPath,
		bazPath,
		fee500,
		-887220,
		887220,
		"50000",
		"50000",
		"0",
		"0",
		max_timeout,
		admin,
		admin,
		"",
	)
}
