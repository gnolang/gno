package mytoken

import (
	"gno.land/p/demo/grc/grc20"
	"std"
)

var (
	mytoken *grc20.AdminToken
	admin   std.Address
)

// init is called once at time of deployment
func init() {
	// set admin as deployer
	admin = std.PrevRealm().Addr()

	// provision the token's name, symbol and number of decimals
	mytoken = grc20.NewAdminToken("My Token", "TKN", 4)

	// mint 1 million tokens to admin
	mytoken.Mint(admin, 1000000*10000)
}
