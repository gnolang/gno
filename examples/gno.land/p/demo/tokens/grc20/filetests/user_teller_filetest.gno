// PKGPATH: gno.land/r/demo/grc20usertoken

// UserTeller accepts direct MsgCall users and rejects realm callers.
package grc20usertoken

import (
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/testutils/v0"
	"gno.land/r/tests/grc20_user_teller_leak"
)

var (
	token  *grc20.Token
	ledger *grc20.PrivateLedger
	teller grc20.Teller
)

func init(cur realm) {
	token, ledger = grc20.NewToken("UserToken", "USER", 4, 0, cur)
	teller = ledger.UserTeller()
}

func Transfer(cur realm, to address, amount int64) error {
	return teller.Transfer(0, cur, to, amount)
}

func Approve(cur realm, spender address, amount int64) error {
	return teller.Approve(0, cur, spender, amount)
}

func TransferFrom(cur realm, owner, to address, amount int64) error {
	return teller.TransferFrom(0, cur, owner, to, amount)
}

func main(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	spender := testutils.TestAddress("spender")
	carol := testutils.TestAddress("carol")
	ledger.Mint(alice, 1_000)

	grc20_user_teller_leak.Store(cross(cur), teller)

	testing.SetRealm(testing.NewUserRealm(alice))
	foreignErr := grc20_user_teller_leak.Transfer(cross(cur), bob, 1)
	println("foreign teller:", foreignErr.Error())
	if err := Transfer(cross(cur), bob, 100); err != nil {
		panic("direct EOA transfer rejected: " + err.Error())
	}
	if err := Approve(cross(cur), spender, 300); err != nil {
		panic("direct EOA approval rejected: " + err.Error())
	}

	testing.SetRealm(testing.NewUserRealm(spender))
	if err := TransferFrom(cross(cur), alice, carol, 125); err != nil {
		panic("direct EOA transferFrom rejected: " + err.Error())
	}
	println("eoa balances:", token.BalanceOf(alice), token.BalanceOf(bob), token.BalanceOf(carol))
	println("remaining allowance:", token.Allowance(alice, spender))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/relay"))
	approveErr := Approve(cross(cur), spender, 1)
	transferFromErr := TransferFrom(cross(cur), alice, carol, 1)
	println("realm approve:", approveErr.Error())
	println("realm transferFrom:", transferFromErr.Error())
	println("unchanged allowance:", token.Allowance(alice, spender))

	println("final balances:", token.BalanceOf(alice), token.BalanceOf(bob), token.BalanceOf(carol))
}

// Output:
// foreign teller: caller teller is confined to the token's own realm
// eoa balances: 775 100 125
// remaining allowance: 175
// realm approve: user teller requires a direct EOA call (MsgCall)
// realm transferFrom: user teller requires a direct EOA call (MsgCall)
// unchanged allowance: 175
// final balances: 775 100 125
