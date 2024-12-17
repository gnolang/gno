package tori

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/users"
	rusers "gno.land/r/demo/users"
)

var (
	tori     *grc20.AdminToken
	userTori grc20.IGRC20
	admin    std.Address = std.DerivePkgAddr("gno.land/r/teritori/dao_realm")
)

func init() {
	tori = grc20.NewAdminToken("Tori", "TORI", 6)
	userTori = tori.GRC20()
}

// method proxies as public functions.
//

// getters.

func TotalSupply() uint64 {
	return tori.TotalSupply()
}

func BalanceOf(owner users.AddressOrName) uint64 {
	balance, err := tori.BalanceOf(rusers.Resolve(owner))
	if err != nil {
		panic(err)
	}
	return balance
}

func Allowance(owner, spender users.AddressOrName) uint64 {
	allowance, err := tori.Allowance(rusers.Resolve(owner), rusers.Resolve(spender))
	if err != nil {
		panic(err)
	}
	return allowance
}

// setters.

func Transfer(to users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	tori.Transfer(caller, rusers.Resolve(to), amount)
}

func Approve(spender users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	tori.Approve(caller, rusers.Resolve(spender), amount)
}

func TransferFrom(from, to users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	tori.TransferFrom(caller, rusers.Resolve(from), rusers.Resolve(to), amount)
}

// administration.

func ChangeAdmin(newAdmin users.AddressOrName) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	admin = rusers.Resolve(newAdmin)
}

func Mint(address users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	tori.Mint(rusers.Resolve(address), amount)
}

func Burn(address users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	tori.Burn(rusers.Resolve(address), amount)
}

// render.
//

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return tori.RenderHome()

	case c == 2 && parts[0] == "balance":
		owner := users.AddressOrName(parts[1])
		balance, _ := tori.BalanceOf(rusers.Resolve(owner))
		return ufmt.Sprintf("%d\n", balance)

	default:
		return "404\n"
	}
}

func assertIsAdmin(address std.Address) {
	if address != admin {
		panic("restricted access")
	}
}
