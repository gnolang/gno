package foo20

import (
	"std"
	"strings"

	grc20 "gno.land/p/grc/grc20/impl"
	"gno.land/p/ufmt"
)

var foo *grc20.Token
var admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin

func init() {
	foo = grc20.NewToken("Foo", "FOO", 4)
	foo.Mint("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", 1000000*10000) // @administrator (1M)
	foo.Mint("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", 10000*10000)   // @manfred (10k)
}

// method proxies as public functions.
//

// getters.

func TotalSupply() uint64                    { return foo.TotalSupply() }
func BalanceOf(address string) uint64        { return foo.BalanceOf(address) }
func Allowance(owner, spender string) uint64 { return foo.Allowance(owner, spender) }

// setters.

func Transfer(to string, amount uint64) {
	caller := std.GetCallerAt(2).String()
	foo.Transfer(caller, to, amount)
}

func Approve(spender string, amount uint64) {
	caller := std.GetCallerAt(2).String()
	foo.Approve(caller, spender, amount)
}

func TransferFrom(from, to string, amount uint64) {
	caller := std.GetCallerAt(2).String()
	foo.TransferFrom(caller, from, to, amount)
}

// administration.

func Mint(address string, amount uint64) {
	caller := std.GetCallerAt(2)
	assertIsAdmin(caller)
	foo.Mint(address, amount)
}

func Burn(address string, amount uint64) {
	caller := std.GetCallerAt(2)
	assertIsAdmin(caller)
	foo.Burn(address, amount)
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
		nameOrAddr := parts[1]
		balance := foo.BalanceOf(nameOrAddr)
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
