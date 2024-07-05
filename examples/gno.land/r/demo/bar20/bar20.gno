// Package bar20 is similar to foo20 but exposes a safe-object that can be used
// by `maketx run`, another contract importing foo20, and in the future when
// we'll support `maketx call Token.XXX`.
package bar20

import (
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
)

var (
	banker *grc20.AdminToken // private banker.
	Token  grc20.IGRC20      // public safe-object.
)

func init() {
	banker = grc20.NewAdminToken("Bar", "BAR", 4)
	Token = banker.GRC20()
	grc20reg.Register(Token, "")
}

func Faucet() string {
	caller := std.PrevRealm().Addr()
	if err := banker.Mint(caller, 1_000_000); err != nil {
		return "error: " + err.Error()
	}
	return "OK"
}

func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return banker.RenderHome() // XXX: should be Token.RenderHome()
	case c == 2 && parts[0] == "balance":
		owner := std.Address(parts[1])
		balance, _ := Token.BalanceOf(owner)
		return ufmt.Sprintf("%d\n", balance)
	default:
		return "404\n"
	}
}
