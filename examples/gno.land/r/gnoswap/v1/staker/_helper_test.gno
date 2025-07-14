package staker

import (
	"math"
	"std"
	"testing"

	"gno.land/p/demo/testutils"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/common"
	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	"gno.land/r/onbloc/foo"
	"gno.land/r/onbloc/obl"
	"gno.land/r/onbloc/qux"
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
	maxApprove  int64  = 9223372036854775806
	max_timeout int64  = 9999999999
)

var (
	// define addresses to use in tests
	addr01 = testutils.TestAddress("addr01")
	addr02 = testutils.TestAddress("addr02")
)

var (
	adminAddr, _ = access.GetAddress(access.ROLE_ADMIN)
	admin        = adminAddr
	adminRealm   = std.NewUserRealm(admin)

	alice = testutils.TestAddress("alice")

	govGovernanceAddr, _ = access.GetAddress(access.ROLE_GOVERNANCE)
	poolAddr, _          = access.GetAddress(access.ROLE_POOL)
	routerAddr, _        = access.GetAddress(access.ROLE_ROUTER)
	emissionAddr, _      = access.GetAddress(access.ROLE_EMISSION)

	posRealm    = std.NewCodeRealm(consts.POSITION_PATH)
	rouRealm    = std.NewCodeRealm(consts.ROUTER_PATH)
	stakerRealm = std.NewCodeRealm(consts.STAKER_PATH)

	// addresses used in tests
	addrUsedInTest = []std.Address{addr01, addr02}

	dummyAddr  = testutils.TestAddress("dummy")
	dummyRealm = std.NewUserRealm(dummyAddr)
)

func CreatePool(t *testing.T,
	token0 string,
	token1 string,
	fee uint32,
	sqrtPriceX96 string,
	caller std.Address,
) {
	t.Helper()

	testing.SetRealm(std.NewUserRealm(caller))
	poolPath := pl.GetPoolPath(token0, token1, fee)
	if !pl.DoesPoolPathExist(poolPath) {
		pl.CreatePool(cross, token0, token1, fee, sqrtPriceX96)
	}
}

func LPTokenStake(t *testing.T, owner std.Address, positionId uint64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
}

func LPTokenUnStake(t *testing.T, owner std.Address, positionId uint64, unwrap bool) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
}

func CreateSecondPoolWithoutFee(t *testing.T) {
	testing.SetRealm(adminRealm)
	pl.SetPoolCreationFeeByAdmin(cross, 0)

	CreatePool(t,
		barPath,
		bazPath,
		fee3000,
		common.TickMathGetSqrtRatioAtTick(0).ToString(),
		admin,
	)
}

func MakeMintPositionWithoutFee(t *testing.T) (uint64, string, string, string) {
	t.Helper()

	// make actual data to test resetting not only position's state but also pool's state
	testing.SetRealm(adminRealm)

	TokenApprove(t, barPath, admin, poolAddr, maxApprove)
	TokenApprove(t, bazPath, admin, poolAddr, maxApprove)

	// mint position
	return pn.Mint(
		cross,
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

func TokenFaucet(t *testing.T, tokenPath string, to std.Address) {
	t.Helper()
	testing.SetOriginCaller(admin)
	defaultAmount := int64(5_000_000_000)

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

func TokenBalance(t *testing.T, tokenPath string, owner std.Address) int64 {
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

func TokenAllowance(t *testing.T, tokenPath string, owner, spender std.Address) int64 {
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

func TokenApprove(t *testing.T, tokenPath string, owner, spender std.Address, amount int64) {
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

func MintPosition(t *testing.T,
	token0 string,
	token1 string,
	fee uint32,
	tickLower int32,
	tickUpper int32,
	amount0Desired string, // *u256.Uint
	amount1Desired string, // *u256.Uint
	amount0Min string, // *u256.Uint
	amount1Min string, // *u256.Uint
	deadline int64,
	mintTo std.Address,
	caller std.Address,
) (uint64, string, string, string) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(caller))

	return pn.Mint(
		cross,
		token0,
		token1,
		fee,
		tickLower,
		tickUpper,
		amount0Desired,
		amount1Desired,
		amount0Min,
		amount1Min,
		deadline,
		mintTo,
		caller,
		"",
	)
}

func wugnotApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	wugnot.Approve(cross, spender, amount)
}

func gnsApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	gns.Approve(cross, spender, amount)
}

func barApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	bar.Approve(cross, spender, amount)
}

func bazApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	baz.Approve(cross, spender, amount)
}

func fooApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	foo.Approve(cross, spender, amount)
}

func oblApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	obl.Approve(cross, spender, amount)
}

func quxApprove(t *testing.T, owner, spender std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(owner))
	qux.Approve(cross, spender, amount)
}

func wugnotTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	wugnot.Transfer(cross, to, amount)
}

func gnsTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	gns.Transfer(cross, to, amount)
}

func barTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	bar.Transfer(cross, to, amount)
}

func bazTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	baz.Transfer(cross, to, amount)
}

func fooTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	foo.Transfer(cross, to, amount)
}

func oblTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	obl.Transfer(cross, to, amount)
}

func quxTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(admin))
	qux.Transfer(cross, to, amount)
}

// ----------------------------------------------------------------------------
// ugnot

func ugnotTransfer(t *testing.T, from, to std.Address, amount int64) {
	t.Helper()

	testing.SetRealm(std.NewUserRealm(from))
	testing.SetOriginSend(std.Coins{{ugnotDenom, amount}})
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(from, to, std.Coins{{ugnotDenom, amount}})
}

func ugnotBalanceOf(t *testing.T, addr std.Address) int64 {
	t.Helper()

	banker := std.NewBanker(std.BankerTypeRealmIssue)
	coins := banker.GetCoins(addr)
	if len(coins) == 0 {
		return 0
	}

	return coins.AmountOf(ugnotDenom)
}

func ugnotMint(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	testing.IssueCoins(addr, std.Coins{{denom, amount}})
}

func ugnotBurn(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	banker := std.NewBanker(std.BankerTypeRealmIssue)
	banker.RemoveCoin(addr, denom, amount)
}

func ugnotFaucet(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	faucetAddress := admin
	testing.SetOriginCaller(faucetAddress)

	if ugnotBalanceOf(t, faucetAddress) < amount {
		newCoins := std.Coins{{ugnotDenom, amount}}
		ugnotMint(t, faucetAddress, newCoins[0].Denom, newCoins[0].Amount)
		testing.SetOriginSend(newCoins)
	}
	ugnotTransfer(t, faucetAddress, to, amount)
}

func ugnotDeposit(t *testing.T, addr std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(addr))
	wugnotAddr := consts.WUGNOT_ADDR
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(addr, wugnotAddr, std.Coins{{ugnotDenom, amount}})
	wugnot.Deposit(cross)
}

func deletePoolTier(t *testing.T, poolPath string) {
	t.Helper()
	if poolTier != nil {
		poolTier.changeTier(std.ChainHeight(), pools, poolPath, 0)
	} else {
		panic("poolTier is nil")
	}
}

func changeWarmup(t *testing.T, index int, blockDuration int64) {
	modifyWarmup(index, blockDuration)
}

type gnsBalanceTracker struct {
	height               int64
	stakerBalance        uint64
	devOpsBalance        uint64
	communityPoolBalance uint64
	govStakerBalance     uint64
	protocolFeeBalance   uint64
	callerBalance        uint64
}

// returns true if actual is within 0.0001% of expected
func isInErrorRange(expected uint64, actual uint64) bool {
	maxSafeValue := uint64(math.MaxUint64 / 100001)
	var lowerBound, upperBound uint64
	if expected > maxSafeValue {
		lowerBound = expected / 1000000 * 999999
		upperBound = expected / 1000000 * 1000001
	} else {
		lowerBound = expected * 999999 / 1000000
		upperBound = expected * 1000001 / 1000000
	}
	return actual >= lowerBound && actual <= upperBound
}
