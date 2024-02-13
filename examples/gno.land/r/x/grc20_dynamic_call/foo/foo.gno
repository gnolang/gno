package foo

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

var (
	foo   *grc20.AdminToken
	admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin
)

func init() {
	foo = grc20.NewAdminToken("Foo", "FOO", 4)
	foo.Mint(admin, 1000000*10000)                                    // @administrator (1M)
	foo.Mint("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", 10000*10000) // @manfred (10k)

	foo.Mint("g1tkmrcu9m0xjddxh0c29r3zac8m2yjwaytlvh6u", 1000*10000) // REALM_gno.land/r/x/grc20_dynamic_call/registry
}

// method proxies as public functions.
//

// getters.

func TotalSupply() uint64 {
	return foo.TotalSupply()
}

func BalanceOf(owner users.AddressOrName) uint64 {
	balance, err := foo.BalanceOf(owner.Resolve())
	if err != nil {
		panic(err.Error())
	}
	return balance
}

func Allowance(owner, spender users.AddressOrName) uint64 {
	allowance, err := foo.Allowance(owner.Resolve(), spender.Resolve())
	if err != nil {
		panic(err.Error())
	}
	return allowance
}

// setters.

func Transfer(to users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := foo.Transfer(caller, to.Resolve(), amount)
	if err != nil {
		panic(err.Error())
	}
}

func Approve(spender users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := foo.Approve(caller, spender.Resolve(), amount)
	if err != nil {
		panic(err.Error())
	}
}

func TransferFrom(from, to users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	err := foo.TransferFrom(caller, from.Resolve(), to.Resolve(), amount)
	if err != nil {
		panic(err.Error())
	}
}

// faucet.

func Faucet() {
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PrevRealm().Addr()
	err := foo.Mint(caller, 1000*10000) // 1k
	if err != nil {
		panic(err.Error())
	}
}

// administration.

func Mint(address users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := foo.Mint(address.Resolve(), amount)
	if err != nil {
		panic(err.Error())
	}
}

func Burn(address users.AddressOrName, amount uint64) {
	caller := std.PrevRealm().Addr()
	assertIsAdmin(caller)
	err := foo.Burn(address.Resolve(), amount)
	if err != nil {
		panic(err.Error())
	}
}

// render.
//

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return foo.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := users.AddressOrName(parts[1])
		balance, _ := foo.BalanceOf(owner.Resolve())
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
