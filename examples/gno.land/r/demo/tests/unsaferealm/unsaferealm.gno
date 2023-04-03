package unsaferealm

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	foo *grc20.AdminToken
	Owner = std.Address("g1lpnflsxpr84dsqkznw85yd5wdzenkj89vsptmf")
)

func init() {
	foo = grc20.NewAdminToken("Fake", "FAKE", 4)

	// std.TestDerivePkgAddr("gno.land/r/demo/tests/unsaferealm")
	// foo.Mint("g1lpnflsxpr84dsqkznw85yd5wdzenkj89vsptmf", 1000000*10000)
	foo.Mint(Owner, 1000000*10000)
	// foo.Mint(std.GetOrigPkgAddr(), 1000000*10000)
}

/*
** Some grc20 functions
 */

func GetCaller() std.Address {
	return std.GetCallerAt(1)
}

func BalanceOf(owner std.Address) uint64 {
	balance, err := foo.BalanceOf(owner)
	if err != nil {
		panic(err)
	}

	return balance
}

func Transfer(to std.Address, amount uint64) {
	caller := std.GetCaller()
	println("transfering", amount, "from:", caller, "to:", to)
	foo.Transfer(caller, to, amount)
}

/*
** Realm unsafe functions
 */

func Do(fn func()) {
	fn()
}
