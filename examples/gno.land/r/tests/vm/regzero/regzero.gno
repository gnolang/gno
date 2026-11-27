package regzero

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
	Token, led = grc20.NewToken("Zero", "ZRO", 4, 0, cur)
}

// TryRegister offers foo20's Approve in the transfer slot.
func TryRegister(cur realm) {
	grc20reg.Register(cross(cur), Token, "", foo20.Approve, Approve, TransferFrom)
}

func Approve(cur realm, spender address, amount int64) {
	if err := led.CallerTeller().Approve(0, cur, spender, amount); err != nil {
		panic(err)
	}
}

func TransferFrom(cur realm, from, to address, amount int64) {
	if err := led.CallerTeller().TransferFrom(0, cur, from, to, amount); err != nil {
		panic(err)
	}
}
