// foo20 is a GRC20 token contract where all the GRC20 methods are proxified
// with top-level functions. see also gno.land/r/demo/bar20.
package foo20

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/grc20reg"
	"gno.land/r/demo/users"
)

var (
	Bank, adm  = grc20.NewBank("Foo", "FOO", 4)
	UserBanker = grc20.PrevRealmBanker(Bank)
	owner      = ownable.NewWithAddress("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq") // @manfred
)

func init() {
	adm.Mint(owner.Owner(), 1_000_000*10_000) // @administrator (1M)
	grc20reg.Register(Bank, "")
}

func TotalSupply() uint64 {
	return UserBanker.TotalSupply()
}

func BalanceOf(owner pusers.AddressOrName) uint64 {
	ownerAddr := users.Resolve(owner)
	return UserBanker.BalanceOf(ownerAddr)
}

func Allowance(owner, spender pusers.AddressOrName) uint64 {
	ownerAddr := users.Resolve(owner)
	spenderAddr := users.Resolve(spender)
	return UserBanker.Allowance(ownerAddr, spenderAddr)
}

func Transfer(to pusers.AddressOrName, amount uint64) {
	toAddr := users.Resolve(to)
	checkErr(UserBanker.Transfer(toAddr, amount))
}

func Approve(spender pusers.AddressOrName, amount uint64) {
	spenderAddr := users.Resolve(spender)
	checkErr(UserBanker.Approve(spenderAddr, amount))
}

func TransferFrom(from, to pusers.AddressOrName, amount uint64) {
	fromAddr := users.Resolve(from)
	toAddr := users.Resolve(to)
	checkErr(UserBanker.TransferFrom(fromAddr, toAddr, amount))
}

// Faucet is distributing foo20 tokens without restriction (unsafe).
// For a real token faucet, you should take care of setting limits are asking payment.
func Faucet() {
	caller := std.PrevRealm().Addr()
	amount := uint64(1_000 * 10_000) // 1k
	checkErr(adm.Mint(caller, amount))
}

func Mint(to pusers.AddressOrName, amount uint64) {
	owner.AssertCallerIsOwner()
	toAddr := users.Resolve(to)
	checkErr(adm.Mint(toAddr, amount))
}

func Burn(from pusers.AddressOrName, amount uint64) {
	owner.AssertCallerIsOwner()
	fromAddr := users.Resolve(from)
	checkErr(adm.Burn(fromAddr, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return Bank.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := pusers.AddressOrName(parts[1])
		ownerAddr := users.Resolve(owner)
		balance := UserBanker.BalanceOf(ownerAddr)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
