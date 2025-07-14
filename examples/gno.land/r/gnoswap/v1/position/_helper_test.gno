package position

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gnft"

	pl "gno.land/r/gnoswap/v1/pool"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	"gno.land/r/onbloc/foo"
	"gno.land/r/onbloc/obl"
	"gno.land/r/onbloc/qux"
	"gno.land/r/onbloc/usdc"

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

	fee100               uint32 = 100
	fee500               uint32 = 500
	fee3000              uint32 = 3000
	fee10000             uint32 = 10000
	maxApprove           int64  = 9223372036854775806
	max_timeout          int64  = 9999999999
	maxSqrtPriceLimitX96 string = "1461446703485210103287273052203988822378723970341"

	TIER_1 uint64 = 1
	TIER_2 uint64 = 2
	TIER_3 uint64 = 3
)

var (
	// define addresses to use in tests
	addr01 = testutils.TestAddress("addr01")
	addr02 = testutils.TestAddress("addr02")
)

var (
	adminAddr, _ = access.GetAddress(access.ROLE_ADMIN)

	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")

	emissionAddr, _    = access.GetAddress(access.ROLE_EMISSION)
	poolAddr, _        = access.GetAddress(access.ROLE_POOL)
	protocolFeeAddr, _ = access.GetAddress(access.ROLE_PROTOCOL_FEE)
	routerAddr, _      = access.GetAddress(access.ROLE_ROUTER)
	devOpsAddr, _      = access.GetAddress(access.ROLE_DEVOPS)

	adminRealm = std.NewUserRealm(adminAddr)
	posRealm   = std.NewCodeRealm(consts.POSITION_PATH)
	rouRealm   = std.NewCodeRealm(consts.ROUTER_PATH)

	// addresses used in tests
	addrUsedInTest = []std.Address{addr01, addr02}
)

func InitialisePoolTest(t *testing.T) {
	t.Helper()

	ugnotFaucet(t, adminAddr, 100_000_000_000_000)
	ugnotDeposit(t, adminAddr, 100_000_000_000_000)

	testing.SetOriginCaller(adminAddr)
	TokenApprove(t, gnsPath, adminAddr, poolAddr, maxApprove)
	poolPath := pl.GetPoolPath(wugnotPath, gnsPath, fee3000)
	if !pl.DoesPoolPathExist(poolPath) {
		pl.CreatePool(cross, wugnotPath, gnsPath, fee3000, "79228162514264337593543950336")
	}

	// 2. create position
	testing.SetOriginCaller(alice)
	TokenFaucet(t, wugnotPath, alice)
	TokenFaucet(t, gnsPath, alice)
	TokenApprove(t, wugnotPath, alice, poolAddr, int64(1000))
	TokenApprove(t, gnsPath, alice, poolAddr, int64(1000))
	MintPosition(t,
		wugnotPath,
		gnsPath,
		fee3000,
		int32(1020),
		int32(5040),
		"1000",
		"1000",
		"0",
		"0",
		max_timeout,
		alice,
		alice,
	)
}

func TokenFaucet(t *testing.T, tokenPath string, to std.Address) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)
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

	return Mint(
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

func MintPositionAll(t *testing.T, caller std.Address) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(caller))
	TokenApprove(t, gnsPath, caller, poolAddr, maxApprove)
	TokenApprove(t, gnsPath, caller, routerAddr, maxApprove)
	TokenApprove(t, wugnotPath, caller, poolAddr, maxApprove)
	TokenApprove(t, wugnotPath, caller, routerAddr, maxApprove)

	params := []struct {
		tickLower int32
		tickUpper int32
		liquidity uint64
		zeroToOne bool
	}{
		{
			tickLower: -300,
			tickUpper: -240,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -240,
			tickUpper: -180,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -180,
			tickUpper: -120,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -120,
			tickUpper: -60,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -60,
			tickUpper: 0,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: 0,
			tickUpper: 60,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 60,
			tickUpper: 120,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 120,
			tickUpper: 180,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 180,
			tickUpper: 240,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 240,
			tickUpper: 300,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: -360,
			tickUpper: -300,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -420,
			tickUpper: -360,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -480,
			tickUpper: -420,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -540,
			tickUpper: -480,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: -600,
			tickUpper: -540,
			liquidity: 10,
			zeroToOne: true,
		},
		{
			tickLower: 300,
			tickUpper: 360,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 360,
			tickUpper: 420,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 420,
			tickUpper: 480,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 480,
			tickUpper: 540,
			liquidity: 10,
			zeroToOne: false,
		},
		{
			tickLower: 540,
			tickUpper: 600,
			liquidity: 10,
			zeroToOne: false,
		},
	}

	for _, p := range params {
		MintPosition(t,
			wugnotPath,
			gnsPath,
			fee3000,
			p.tickLower,
			p.tickUpper,
			"100",
			"100",
			"0",
			"0",
			max_timeout,
			caller,
			caller)
	}
}

// set pool create fee to 0 for testing
func CreatePoolWithoutFee(t *testing.T) {
	testing.SetRealm(adminRealm)
	pl.SetPoolCreationFeeByAdmin(cross, 0)
	CreatePool(t, barPath, fooPath, fee500, common.TickMathGetSqrtRatioAtTick(0).ToString(), adminAddr)
}

func MakeMintPositionWithoutFee(t *testing.T) (uint64, string, string, string) {
	t.Helper()

	// make actual data to test resetting not only position's state but also pool's state
	testing.SetRealm(adminRealm)

	// set pool create fee to 0 for testing
	pl.SetPoolCreationFeeByAdmin(cross, 0)
	CreatePool(t, barPath, fooPath, fee500, common.TickMathGetSqrtRatioAtTick(0).ToString(), adminAddr)

	TokenApprove(t, barPath, adminAddr, poolAddr, maxApprove)
	TokenApprove(t, fooPath, adminAddr, poolAddr, maxApprove)

	// mint position
	return Mint(
		cross,
		barPath,
		fooPath,
		fee500,
		-887270,
		887270,
		"50000",
		"50000",
		"0",
		"0",
		max_timeout,
		adminAddr,
		adminAddr,
		"",
	)
}

func LPTokenApprove(t *testing.T, owner, operator std.Address, positionId uint64) {
	t.Helper()
	testing.SetOriginCaller(owner)
	gnft.Approve(cross, operator, positionIdFrom(positionId))
}

// func LPTokenStake(t *testing.T, owner std.Address, positionId uint64) {
// 	t.Helper()
// 	testing.SetRealm(std.NewUserRealm(owner))
// 	sr.StakeToken(positionId, "")
// }

// func LPTokenUnStake(t *testing.T, owner std.Address, positionId uint64, unwrap bool) {
// 	t.Helper()
// 	testing.SetRealm(std.NewUserRealm(owner))
// 	sr.UnStakeToken(positionId, unwrap)
// }

func getPoolFromPositionId(t *testing.T, positionId uint64) *pl.Pool {
	t.Helper()
	position := MustGetPosition(positionId)
	return pl.GetPoolFromPoolPath(position.poolKey)
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
	testing.SetRealm(std.NewUserRealm(adminAddr))
	wugnot.Transfer(cross, to, amount)
}

func gnsTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(adminAddr))
	gns.Transfer(cross, to, amount)
}

func barTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(adminAddr))
	bar.Transfer(cross, to, amount)
}

func bazTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(adminAddr))
	baz.Transfer(cross, to, amount)
}

func fooTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(adminAddr))
	foo.Transfer(cross, to, amount)
}

func oblTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(adminAddr))
	obl.Transfer(cross, to, amount)
}

func quxTransfer(t *testing.T, to std.Address, amount int64) {
	t.Helper()
	testing.SetRealm(std.NewUserRealm(adminAddr))
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
	faucetAddress := adminAddr
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

// resetObject resets the object state(clear or make it default values)
func resetObject(t *testing.T) {
	positions = avl.NewTree()
	nextId = 1
}

func burnTokens(t *testing.T) {
	t.Helper()

	// burn tokens
	for _, addr := range addrUsedInTest {
		burnFoo(addr)
		burnBar(addr)
		burnBaz(addr)
		burnQux(addr)
		burnObl(addr)
		burnUsdc(addr)
	}
}

func burnFoo(addr std.Address) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	foo.Burn(cross, addr, foo.BalanceOf(addr))
}

func burnBar(addr std.Address) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	bar.Burn(cross, addr, bar.BalanceOf(addr))
}

func burnBaz(addr std.Address) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	baz.Burn(cross, addr, baz.BalanceOf(addr))
}

func burnQux(addr std.Address) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	qux.Burn(cross, addr, qux.BalanceOf(addr))
}

func burnObl(addr std.Address) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	obl.Burn(cross, addr, obl.BalanceOf(addr))
}

func burnUsdc(addr std.Address) {
	testing.SetRealm(std.NewUserRealm(adminAddr))
	usdc.Burn(cross, addr, usdc.BalanceOf(addr))
}

// burnAllNFT burns all NFTs
func burnAllNFT(t *testing.T) {
	t.Helper()

	testing.SetRealm(std.NewCodeRealm(consts.POSITION_PATH))
	for i := int64(1); i <= gnft.TotalSupply(); i++ {
		gnft.Burn(cross, positionIdFrom(i))
	}
}

func TestBeforeResetPositionObject(t *testing.T) {
	t.Skip("only works when testing this single file")
	// make actual data to test resetting not only position's state but also pool's state
	testing.SetRealm(adminRealm)

	positionId, liquidity, amount0, amount1 := MakeMintPositionWithoutFee(t)
	uassert.Equal(t, positionId, uint64(1), "positionId should be 1")
	uassert.Equal(t, liquidity, "50000", "liquidity should be 50000")
	uassert.Equal(t, amount0, "50000", "amount0 should be 50000")
	uassert.Equal(t, amount1, "50000", "amount1 should be 50000")
	uassert.Equal(t, positions.Size(), 1, "positions should have 1 position")
	uassert.Equal(t, nextId, uint64(2), "nextId should be 2")
	uassert.Equal(t, gnft.TotalSupply(), int64(1), "gnft total supply should be 1")
	uassert.Equal(t, pl.PoolGetLiquidity("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500"), "50000", "pool liquidity should be 50000")
}

func TestResetObject(t *testing.T) {
	t.Skip("only works when testing this single file")
	resetObject(t)

	uassert.Equal(t, positions.Size(), 0, "positions should be empty")
	uassert.Equal(t, nextId, uint64(1), "nextId should be 1")
}

func TestBurnTokens(t *testing.T) {
	t.Skip("only works when testing this single file")
	burnTokens(t)

	uassert.Equal(t, foo.BalanceOf(addr01), int64(0)) // 100_000_000 -> 0
	uassert.Equal(t, bar.BalanceOf(addr01), int64(0)) // 100_000_000 -> 0
}

func TestBurnAllNFT(t *testing.T) {
	t.Skip("only works when testing this single file")
	burnAllNFT(t)
	uassert.Equal(t, gnft.TotalSupply(), int64(0), "gnft total supply should be 0")
}
