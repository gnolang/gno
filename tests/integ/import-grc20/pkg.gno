package importgrc20

import (
	"std"
	"strings"

	"gno.land/p/grc/grc20"
	"gno.land/r/users"
)

var (
	foo   *grc20.AdminToken
	admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
)

func init() {
	foo = grc20.NewAdminToken("Foo", "FOO", 4)
	foo.Mint(admin, 1000000*10000)
}

func BalanceOf(owner users.AddressOrName) uint64 {
	balance, err := foo.BalanceOf(owner.Resolve())
	if err != nil {
		panic(err)
	}
	return balance
}
