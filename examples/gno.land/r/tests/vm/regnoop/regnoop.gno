package regnoop

import (
	"gno.land/p/nt/grc20/v0"
	"gno.land/r/nt/grc20reg/v0"
)

var Token *grc20.Token

func init(cur realm) {
	Token, _ = grc20.NewToken("Noop", "NOP", 4, 0, cur)
}

// TryRegister registers this realm's own functions, which do nothing. The
// registry vouches for whose code runs, so this registration is accepted.
func TryRegister(cur realm) {
	grc20reg.Register(cross(cur), Token, "", Transfer, Approve, TransferFrom)
}

func Transfer(cur realm, to address, amount int64)           {}
func Approve(cur realm, spender address, amount int64)       {}
func TransferFrom(cur realm, from, to address, amount int64) {}
