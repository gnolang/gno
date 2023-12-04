package mytoken

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	mytoken *grc20.AdminToken
	admin   std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // set admin account
)

// init is a constructor function that runs only once (at time of deployment)
func init() {
	// provision the token's name, symbol and number of decimals
	mytoken = grc20.NewAdminToken("Mytoken", "MTKN", 4)

	// set the total supply
	mytoken.Mint(admin, 1000000*10000) // @administrator (supply = 1 million)
}
