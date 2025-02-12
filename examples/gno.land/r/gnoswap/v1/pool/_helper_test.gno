package pool

import (
	"std"
	"testing"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
	"gno.land/r/onbloc/foo"
	"gno.land/r/onbloc/obl"
	"gno.land/r/onbloc/qux"
	"gno.land/r/onbloc/usdc"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
	pn "gno.land/r/gnoswap/v1/position"
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

	maxSqrtPriceLimitX96 string = "1461446703485210103287273052203988822378723970341"
)

var (
	// define addresses to use in tests
	addr01 = testutils.TestAddress("addr01")
	addr02 = testutils.TestAddress("addr02")
)

var (
	fooToken = common.GetToken(fooPath)
	barToken = common.GetToken(barPath)
	bazToken = common.GetToken(bazPath)
)

var (
	admin       = consts.ADMIN
	adminAddr   = admin
	alice       = testutils.TestAddress("alice")
	pool        = consts.POOL_ADDR
	protocolFee = consts.PROTOCOL_FEE_ADDR
	router      = consts.ROUTER_ADDR
	adminRealm  = std.NewUserRealm(admin)
	posRealm    = std.NewCodeRealm(consts.POSITION_PATH)
	rouRealm    = std.NewCodeRealm(consts.ROUTER_PATH)

	// addresses used in tests
	addrUsedInTest = []std.Address{addr01, addr02}
)

func InitialisePoolTest(t *testing.T) {
	t.Helper()

	ugnotFaucet(t, admin, 100_000_000_000_000)
	ugnotDeposit(t, admin, 100_000_000_000_000)

	std.TestSetOrigCaller(admin)
	TokenApprove(t, gnsPath, admin, pool, maxApprove)
	poolPath := GetPoolPath(wugnotPath, gnsPath, fee3000)
	if !DoesPoolPathExist(poolPath) {
		CreatePool(wugnotPath, gnsPath, fee3000, "79228162514264337593543950336")
	}

	//2. create position
	std.TestSetOrigCaller(alice)
	TokenFaucet(t, wugnotPath, alice)
	TokenFaucet(t, gnsPath, alice)
	TokenApprove(t, wugnotPath, alice, pool, uint64(1000))
	TokenApprove(t, gnsPath, alice, pool, uint64(1000))
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
	std.TestSetOrigCaller(admin)
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
	std.TestSetRealm(std.NewUserRealm(caller))

	return pn.Mint(
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
		caller)
}

func MintPositionAll(t *testing.T, caller std.Address) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(caller))
	TokenApprove(t, gnsPath, caller, pool, maxApprove)
	TokenApprove(t, gnsPath, caller, router, maxApprove)
	TokenApprove(t, wugnotPath, caller, pool, maxApprove)
	TokenApprove(t, wugnotPath, caller, router, maxApprove)

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
	std.TestSetOrigSend(std.Coins{{ugnotDenom, int64(amount)}}, nil)
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(from, to, std.Coins{{ugnotDenom, int64(amount)}})
}

func ugnotBalanceOf(t *testing.T, addr std.Address) uint64 {
	t.Helper()

	banker := std.GetBanker(std.BankerTypeRealmIssue)
	coins := banker.GetCoins(addr)
	if len(coins) == 0 {
		return 0
	}

	return uint64(coins.AmountOf(ugnotDenom))
}

func ugnotMint(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	std.TestIssueCoins(addr, std.Coins{{denom, int64(amount)}})
}

func ugnotBurn(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	banker := std.GetBanker(std.BankerTypeRealmIssue)
	banker.RemoveCoin(addr, denom, amount)
}

func ugnotFaucet(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	faucetAddress := admin
	std.TestSetOrigCaller(faucetAddress)

	if ugnotBalanceOf(t, faucetAddress) < amount {
		newCoins := std.Coins{{ugnotDenom, int64(amount)}}
		ugnotMint(t, faucetAddress, newCoins[0].Denom, newCoins[0].Amount)
		std.TestSetOrigSend(newCoins, nil)
	}
	ugnotTransfer(t, faucetAddress, to, amount)
}

func ugnotDeposit(t *testing.T, addr std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(addr))
	wugnotAddr := consts.WUGNOT_ADDR
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(addr, wugnotAddr, std.Coins{{ugnotDenom, int64(amount)}})
	wugnot.Deposit()
}

// resetObject resets the object state(clear or make it default values)
func resetObject(t *testing.T) {
	pools = avl.NewTree()
	slot0FeeProtocol = 0
	poolCreationFee = 100_000_000
	withdrawalFeeBPS = 100
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
	std.TestSetRealm(std.NewUserRealm(admin))
	foo.Burn(addr, foo.BalanceOf(addr))
}

func burnBar(addr std.Address) {
	std.TestSetRealm(std.NewUserRealm(admin))
	bar.Burn(addr, bar.BalanceOf(addr))
}

func burnBaz(addr std.Address) {
	std.TestSetRealm(std.NewUserRealm(admin))
	baz.Burn(addr, baz.BalanceOf(addr))
}

func burnQux(addr std.Address) {
	std.TestSetRealm(std.NewUserRealm(admin))
	qux.Burn(addr, qux.BalanceOf(addr))
}

func burnObl(addr std.Address) {
	std.TestSetRealm(std.NewUserRealm(admin))
	obl.Burn(addr, obl.BalanceOf(addr))
}

func burnUsdc(addr std.Address) {
	std.TestSetRealm(std.NewUserRealm(admin))
	usdc.Burn(addr, usdc.BalanceOf(addr))
}

func TestBeforeResetObject(t *testing.T) {
	// make some data
	pools = avl.NewTree()
	pools.Set("gno.land/r/gnoswap/v1/gns:gno.land/r/onbloc/usdc", &Pool{
		token0Path: "gno.land/r/gnoswap/v1/gns",
		token1Path: "gno.land/r/onbloc/usdc",
	})

	slot0FeeProtocol = 1
	poolCreationFee = 100_000_000
	withdrawalFeeBPS = 100

	// transfer some tokens
	std.TestSetRealm(std.NewUserRealm(admin))
	foo.Transfer(addr01, 100_000_000)
	bar.Transfer(addr01, 100_000_000)

	uassert.Equal(t, foo.BalanceOf(addr01), uint64(100_000_000))
	uassert.Equal(t, bar.BalanceOf(addr01), uint64(100_000_000))
}

func TestResetObject(t *testing.T) {
	resetObject(t)
	uassert.Equal(t, pools.Size(), 0)
	uassert.Equal(t, slot0FeeProtocol, uint8(0))
	uassert.Equal(t, poolCreationFee, uint64(100_000_000))
	uassert.Equal(t, withdrawalFeeBPS, uint64(100))
}

func TestBurnTokens(t *testing.T) {
	burnTokens(t)

	uassert.Equal(t, foo.BalanceOf(addr01), uint64(0)) // 100_000_000 -> 0
	uassert.Equal(t, bar.BalanceOf(addr01), uint64(0)) // 100_000_000 -> 0
}
