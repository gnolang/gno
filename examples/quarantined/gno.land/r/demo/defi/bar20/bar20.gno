// Package bar20 is similar to gno.land/r/demo/foo20.
//
// It used to export its teller value as a "safe object" for `maketx run` and
// for other contracts to call. That is no longer possible, and the reason is
// worth knowing before copying the shape: a teller is a spend capability, and
// the frame-relative one resolves the account to debit from whatever frame
// invokes it. Exported, it debits whoever crossed into the realm that read the
// var — chosen by that realm, not by its owner. grc20 now confines it to the
// token's own realm, so an exported value is inert everywhere else, including
// in a `maketx run` script, whose own pkgpath is not this realm's.
//
// The teller is therefore unexported here and reached only through the narrow
// wrappers below, which is the pattern to copy.
package bar20

import (
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/demo/defi/grc20reg"
)

var (
	Token *grc20.Token
	adm   *grc20.PrivateLedger
	// userTeller stays unexported: it is a spend capability, and it only
	// functions inside this realm. The wrappers below are the public surface.
	userTeller grc20.Teller
)

func init(cur realm) {
	// bar20 only ever creates this one token, so id 0 can't collide.
	Token, adm = grc20.NewToken("Bar", "BAR", 4, 0, cur)
	userTeller = adm.CallerTeller()
	grc20reg.Register(cross(cur), Token, "")
}

// Transfer moves the caller's own tokens. userTeller debits cur.Previous(), so
// a direct user call spends that user's balance and a realm cross-call spends
// the calling realm's.
func Transfer(cur realm, to address, amount int64) {
	checkErr(userTeller.Transfer(0, cur, to, amount))
}

// Approve grants spender an allowance over the caller's balance.
func Approve(cur realm, spender address, amount int64) {
	checkErr(userTeller.Approve(0, cur, spender, amount))
}

// TransferFrom spends an allowance the owner granted to the caller.
func TransferFrom(cur realm, owner, to address, amount int64) {
	checkErr(userTeller.TransferFrom(0, cur, owner, to, amount))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Faucet(cur realm) string {
	caller := cur.Previous().Address()
	if err := adm.Mint(caller, 1_000_000); err != nil {
		return "error: " + err.Error()
	}
	return "OK"
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return Token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := address(parts[1])
		balance := Token.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
