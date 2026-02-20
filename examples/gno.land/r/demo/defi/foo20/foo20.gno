// foo20 is a GRC20 token contract where all the grc20.Teller methods are
// proxified with top-level functions. see also gno.land/r/demo/bar20.
package foo20

import (
	"chain/runtime"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/ufmt"
	"gno.land/r/demo/defi/grc20reg"
)

var (
	Token, privateLedger = grc20.NewToken("Foo", "FOO", 4)
	UserTeller           = Token.CallerTeller()
	Ownable              = ownable.NewWithAddressByPrevious("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @manfred
)

func init() {
	privateLedger.Mint(Ownable.Owner(), 1_000_000*10_000) // @privateLedgeristrator (1M)
	grc20reg.Register(cross, Token, "")
}

func TotalSupply() int64 {
	return UserTeller.TotalSupply()
}

func BalanceOf(owner address) int64 {
	return UserTeller.BalanceOf(owner)
}

func Allowance(owner, spender address) int64 {
	return UserTeller.Allowance(owner, spender)
}

func Transfer(cur realm, to address, amount int64) {
	checkErr(UserTeller.Transfer(to, amount))
}

func Approve(cur realm, spender address, amount int64) {
	checkErr(UserTeller.Approve(spender, amount))
}

func TransferFrom(cur realm, from, to address, amount int64) {
	checkErr(UserTeller.TransferFrom(from, to, amount))
}

// Faucet is distributing foo20 tokens without restriction (unsafe).
// For a real token faucet, you should take care of setting limits are asking payment.
func Faucet(cur realm) {
	caller := runtime.PreviousRealm().Address()
	amount := int64(1_000 * 10_000) // 1k
	checkErr(privateLedger.Mint(caller, amount))
}

func Mint(cur realm, to address, amount int64) {
	Ownable.AssertOwnedByPrevious()
	checkErr(privateLedger.Mint(to, amount))
}

func Burn(cur realm, from address, amount int64) {
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
		owner := address(parts[1])
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
