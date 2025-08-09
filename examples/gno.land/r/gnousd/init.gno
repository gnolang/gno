package gnousd

import (
	"std"

	grc20 "gno.land/p/grc/grc20/impl"
)

var controller *Controller

func init() {
	// TODO: split this into 4 independent packages:
	// - coins autonomous grc20
	// - shares autonomous grc20
	// - controller
	// - oracle placeholder
	adminAddr := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	coins := grc20.NewToken("GnoUSD Coins", "GUSD", 18)
	shares := grc20.NewToken("GnoUSD Shares", "GUSDSH", 18)
	controller = newController(coins, shares, adminAddr)
}
