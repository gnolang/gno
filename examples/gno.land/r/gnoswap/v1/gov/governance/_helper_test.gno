package governance

import (
	"std"

	"gno.land/p/gnoswap/consts"
)

var (
	admin std.Address = consts.ADMIN

	fooPath string = "gno.land/r/onbloc/foo"
	barPath string = "gno.land/r/onbloc/bar"
	bazPath string = "gno.land/r/onbloc/baz"
	quxPath string = "gno.land/r/onbloc/qux"

	oblPath string = "gno.land/r/onbloc/obl"

	fee100  uint32 = 100
	fee500  uint32 = 500
	fee3000 uint32 = 3000

	max_timeout int64 = 9999999999
)

// Realms to mock frames
var (
	adminRealm = std.NewUserRealm(admin)

	posRealm = std.NewCodeRealm(consts.POSITION_PATH)
	rouRealm = std.NewCodeRealm(consts.ROUTER_PATH)
	stkRealm = std.NewCodeRealm(consts.STAKER_PATH)
	govRealm = std.NewCodeRealm(consts.GOV_GOVERNANCE_PATH)
)

/* HELPER */
func ugnotBalanceOf(addr std.Address) uint64 {
	testBanker := std.NewBanker(std.BankerTypeRealmIssue)

	coins := testBanker.GetCoins(addr)
	if len(coins) == 0 {
		return 0
	}

	return uint64(coins.AmountOf("ugnot"))
}
