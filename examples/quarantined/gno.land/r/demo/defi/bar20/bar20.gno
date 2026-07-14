// Package bar20 is similar to gno.land/r/demo/foo20 but exposes a safe-object
// that can be used by `maketx run`, another contract importing foo20, and in
// the future when we'll support `maketx call Token.XXX`.
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
	// UserTeller is intentionally exported for `maketx run` consumption.
	// It is the canonical *fnTeller (verifiable via grc20.IsCanonicalTeller);
	// foreign realms reading this var can call its methods, but the methods
	// resolve the source-of-funds from cur.Previous().Address() via
	// IsCurrent, so a caller can only spend their OWN balance — not the
	// reader's. No cur is leaked via this exposure.
	UserTeller grc20.Teller
)

func init(cur realm) {
	Token, adm = grc20.NewToken("Bar", "BAR", 4, "token", cur)
	UserTeller = Token.CallerTeller()
	grc20reg.Register(cross(cur), Token, "")
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
