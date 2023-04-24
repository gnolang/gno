package fake20

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var (
	foo   *grc20.AdminToken
	admin std.Address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // TODO: helper to change admin
)

func Init() {
	foo = grc20.NewAdminToken("Fake", "FAKE", 4)
	foo.Mint(admin, 1000000*10000)                                    // @administrator (1M)
	foo.Mint("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", 10000*10000) // @manfred (10k)
	foo.Mint("g1lpnflsxpr84dsqkznw85yd5wdzenkj89vsptmf", 10000*10000) // @gno.land/r/demo/tests/unsaferealm
}

func init() {
	Init()
}

// method proxies as public functions.
//

// getters.

func GetCaller() std.Address {
	return std.GetCallerAt(1)
	// return std.GetOrigPkgAddr()
}

func TotalSupply() uint64 {
	return foo.TotalSupply()
}

func BalanceOf(owner std.Address) uint64 {
	balance, err := foo.BalanceOf(owner)
	if err != nil {
		panic(err)
	}
	return balance
}

func Allowance(owner, spender std.Address) uint64 {
	allowance, err := foo.Allowance(owner, spender)
	if err != nil {
		panic(err)
	}
	return allowance
}

// setters.

func Transfer(to std.Address, amount uint64) {
	caller := std.GetCaller()
	println("transfering", amount, "from:", caller, "to:", to)
	foo.Transfer(caller, to, amount)
}

func Approve(spender std.Address, amount uint64) {
	caller := std.GetCaller()
	foo.Approve(caller, spender, amount)
}

func TransferFrom(from, to std.Address, amount uint64) {
	caller := std.GetCaller()
	foo.TransferFrom(caller, from, to, amount)
}

// faucet.

func Faucet() {
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.GetCaller()
	foo.Mint(caller, 1000*10000) // 1k
}

// administration.

func Mint(address std.Address, amount uint64) {
	caller := std.GetCaller()
	assertIsAdmin(caller)
	foo.Mint(address, amount)
}

func Burn(address std.Address, amount uint64) {
	caller := std.GetCaller()
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
		owner := std.Address(parts[1])
		balance, _ := foo.BalanceOf(owner)
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
