package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/wugnot"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/access"
)

var (
	admin, _             = access.GetAddress(access.ROLE_ADMIN)
	stakerAddr, _        = access.GetAddress(access.ROLE_STAKER)
	communityPoolAddr, _ = access.GetAddress(access.ROLE_COMMUNITY_POOL)
	govStakerAddr, _     = access.GetAddress(access.ROLE_GOV_STAKER)
	protocolFeeAddr, _   = access.GetAddress(access.ROLE_PROTOCOL_FEE)
	launchpadAddr, _     = access.GetAddress(access.ROLE_LAUNCHPAD)
	devOpsAddr, _        = access.GetAddress(access.ROLE_DEVOPS)

	fooPath  string = "gno.land/r/onbloc/foo"
	barPath  string = "gno.land/r/onbloc/bar"
	bazPath  string = "gno.land/r/onbloc/baz"
	quxPath  string = "gno.land/r/onbloc/qux"
	oblPath  string = "gno.land/r/onbloc/obl"
	usdcPath string = "gno.land/r/onbloc/usdc"

	ugnotDenom string = "ugnot"
	ugnotPath  string = "ugnot"
	wugnotPath string = "gno.land/r/demo/wugnot"

	fee100  uint32 = 100
	fee500  uint32 = 500
	fee3000 uint32 = 3000

	max_timeout int64 = 9999999999

	realmPrefix = "/gno.land/r/gnoswap/v1/gov/staker"
)

// Realms to mock frames
var (
	adminRealm  = std.NewUserRealm(admin)
	userRealm   = std.NewUserRealm(testutils.TestAddress("alice"))
	user2Realm  = std.NewUserRealm(testutils.TestAddress("bob"))
	user3Realm  = std.NewUserRealm(testutils.TestAddress("charlie"))
	invalidAddr = testutils.TestAddress("invalid")

	posRealm    = std.NewCodeRealm(consts.POSITION_PATH)
	rouRealm    = std.NewCodeRealm(consts.ROUTER_PATH)
	stakerRealm = std.NewCodeRealm(consts.STAKER_PATH)
)

func makeFakeAddress(name string) std.Address {
	return testutils.TestAddress(name)
}

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
	std.TestIssueCoins(addr, std.Coins{{ugnotDenom, int64(amount)}})
}

func ugnotBurn(t *testing.T, addr std.Address, denom string, amount int64) {
	t.Helper()
	banker := std.NewBanker(std.BankerTypeRealmIssue)
	banker.RemoveCoin(addr, ugnotDenom, amount)
}

func ugnotFaucet(t *testing.T, to std.Address, amount uint64) {
	t.Helper()
	// faucet address is admin
	std.TestSetOriginCaller(admin)

	if ugnotBalanceOf(t, admin) < amount {
		ugnotMint(t, admin, ugnotPath, int64(amount))
		std.TestSetOriginSend(std.Coins{{ugnotPath, int64(amount)}}, nil)
	}
	ugnotTransfer(t, admin, to, amount)
}

func ugnotDeposit(t *testing.T, addr std.Address, amount uint64) {
	t.Helper()
	std.TestSetRealm(std.NewUserRealm(addr))
	wugnotAddr := consts.WUGNOT_ADDR
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(addr, wugnotAddr, std.Coins{{ugnotDenom, int64(amount)}})
	wugnot.Deposit()
}
