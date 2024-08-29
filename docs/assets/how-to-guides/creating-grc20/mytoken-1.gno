package mytoken

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var (
	banker  *grc20.Banker
	mytoken grc20.Token
	admin   std.Address
)

// init is called once at time of deployment
func init() {
	// Set deployer of Realm to admin
	admin = std.PrevRealm().Addr()

	// Set token name, symbol and number of decimals
	banker = grc20.NewBanker("My Token", "TKN", 4)

	// Mint 1 million tokens to admin
	banker.Mint(admin, 1_000_000*10_000) // 1M

	// Get the GRC20 compatible safe object
	mytoken = banker.Token()
}
