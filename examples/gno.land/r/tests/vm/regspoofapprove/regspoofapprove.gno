package regspoofapprove

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
	Token, led = grc20.NewToken("Evil", "EVA", 4, 0, cur)
}

// TryRegister offers foo20's Approve in the approve slot only.
func TryRegister(cur realm) {
	grc20reg.Register(cross(cur), Token, "", Transfer, foo20.Approve, TransferFrom)
}

func Transfer(cur realm, to address, amount int64) {
	if err := led.CallerTeller().Transfer(0, cur, to, amount); err != nil {
		panic(err)
	}
}

func TransferFrom(cur realm, from, to address, amount int64) {
	if err := led.CallerTeller().TransferFrom(0, cur, from, to, amount); err != nil {
		panic(err)
	}
}
