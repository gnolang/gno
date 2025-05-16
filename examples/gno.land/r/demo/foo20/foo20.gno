// foo20 is a GRC20 token contract where all the grc20.Teller methods are
// proxified with top-level functions. see also gno.land/r/demo/bar20.
package foo20

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
)

var (
	Token, privateLedger = grc20.NewToken("Foo", "FOO", 4)
	UserTeller           = Token.CallerTeller()
	Ownable              = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @manfred
)

func init() {
	privateLedger.Mint(Ownable.Owner(), 1_000_000*10_000) // @privateLedgeristrator (1M)
	cross(grc20reg.Register)(Token.Getter(), "")
}

func TotalSupply() uint64 {
	return UserTeller.TotalSupply()
}

func BalanceOf(owner std.Address) uint64 {
	return UserTeller.BalanceOf(owner)
}

func Allowance(owner, spender std.Address) uint64 {
	return UserTeller.Allowance(owner, spender)
}

func Transfer(to std.Address, amount uint64) {
	crossing()

	checkErr(UserTeller.Transfer(to, amount))
}

func Approve(spender std.Address, amount uint64) {
	crossing()

	checkErr(UserTeller.Approve(spender, amount))
}

func TransferFrom(from, to std.Address, amount uint64) {
	crossing()

	checkErr(UserTeller.TransferFrom(from, to, amount))
}

// Faucet is distributing foo20 tokens without restriction (unsafe).
// For a real token faucet, you should take care of setting limits are asking payment.
func Faucet() {
	crossing()

	caller := std.PreviousRealm().Address()
	amount := uint64(1_000 * 10_000) // 1k
	checkErr(privateLedger.Mint(caller, amount))
}

func Mint(to std.Address, amount uint64) {
	crossing()

	Ownable.AssertOwnedByPrevious()
	checkErr(privateLedger.Mint(to, amount))
}

func Burn(from std.Address, amount uint64) {
	crossing()

	Ownable.AssertOwnedByPrevious()
	checkErr(privateLedger.Burn(from, amount))
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return Token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance := UserTeller.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
