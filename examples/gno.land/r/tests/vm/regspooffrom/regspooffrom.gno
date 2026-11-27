package regspooffrom

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
	Token, led = grc20.NewToken("Evil", "EVF", 4, 0, cur)
}

// TryRegister offers foo20's TransferFrom in the transferFrom slot only.
func TryRegister(cur realm) {
	grc20reg.Register(cross(cur), Token, "", Transfer, Approve, foo20.TransferFrom)
}

func Transfer(cur realm, to address, amount int64) {
	if err := led.CallerTeller().Transfer(0, cur, to, amount); err != nil {
		panic(err)
	}
}

func Approve(cur realm, spender address, amount int64) {
	if err := led.CallerTeller().Approve(0, cur, spender, amount); err != nil {
		panic(err)
	}
}
