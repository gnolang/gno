// foo20 is a GRC20 token contract where all the grc20.Teller methods are
// proxified with top-level functions. see also gno.land/r/demo/bar20.
//
// It is built on gno.land/p/nt/grc20/v1, so every amount is a uint256, and is
// registered in gno.land/r/nt/grc20reg/v1.
package foo20

import (
	"strings"

	"gno.land/p/nt/grc20/v1"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/nt/grc20reg/v1"
)

var (
	Token         *grc20.Token
	privateLedger *grc20.PrivateLedger
	userTeller    grc20.Teller
	Ownable       = ownable.NewWithAddress("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx") // govdao t1 multisig
)

func init(cur realm) {
	// foo20 only ever creates this one token, so id 0 can't collide.
	Token, privateLedger = grc20.NewToken("Foo", "FOO", 4, 0, cur)
	userTeller = privateLedger.CallerTeller()
	privateLedger.Mint(Ownable.Owner(), 1_000_000*10_000) // @privateLedgeristrator (1M)
	grc20reg.Register(cross(cur), Token, "")
}

func TotalSupply() uint256 {
	return userTeller.TotalSupply()
}

func BalanceOf(owner address) uint256 {
	return userTeller.BalanceOf(owner)
}

func Allowance(owner, spender address) uint256 {
	return userTeller.Allowance(owner, spender)
}

func Transfer(cur realm, to address, amount uint256) {
	checkErr(userTeller.Transfer(0, cur, to, amount))
}

func Approve(cur realm, spender address, amount uint256) {
	checkErr(userTeller.Approve(0, cur, spender, amount))
}

func TransferFrom(cur realm, from, to address, amount uint256) {
	checkErr(userTeller.TransferFrom(0, cur, from, to, amount))
}

// Faucet is distributing foo20 tokens without restriction (unsafe).
// For a real token faucet, you should take care of setting limits are asking payment.
func Faucet(cur realm) {
	caller := cur.Previous().Address()
	amount := uint256(1_000 * 10_000) // 1k
	checkErr(privateLedger.Mint(caller, amount))
}

func Mint(cur realm, to address, amount uint256) {
	Ownable.AssertOwnedBy(cur.Previous().Address())
	checkErr(privateLedger.Mint(to, amount))
}

func Burn(cur realm, from address, amount uint256) {
	Ownable.AssertOwnedBy(cur.Previous().Address())
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
		balance := userTeller.BalanceOf(owner)
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
