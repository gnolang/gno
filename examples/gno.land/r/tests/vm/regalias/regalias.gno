package regalias

import (
	"gno.land/p/nt/grc20/v0"
	"gno.land/r/demo/defi/foo20"
	"gno.land/r/nt/grc20reg/v0"
)

var (
	Token *grc20.Token
	led   *grc20.PrivateLedger
)

func init(cur realm) {
	Token, led = grc20.NewToken("Alias", "ALI", 4, 0, cur)
	grc20reg.Register(cross(cur), Token, "", Transfer, Approve, TransferFrom)
}

// Swap swaps this realm's token contents for foo20's; the registry's copy of
// the struct must keep the key on this ledger.
func Swap(cur realm) {
	*Token = *foo20.Token
}

// The writes go through the ledger value kept here, not through the Token
// struct Swap overwrites.
func Transfer(cur realm, to address, amount int64) {
	if err := led.Transfer(cur.Previous().Address(), to, amount); err != nil {
		panic(err)
	}
}

func Approve(cur realm, spender address, amount int64) {
	if err := led.Approve(cur.Previous().Address(), spender, amount); err != nil {
		panic(err)
	}
}

func TransferFrom(cur realm, from, to address, amount int64) {
	if err := led.TransferFrom(from, cur.Previous().Address(), to, amount); err != nil {
		panic(err)
	}
}
