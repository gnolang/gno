package regspoof

import (
	"gno.land/p/nt/grc20/v0"
	"gno.land/r/demo/defi/foo20"
	"gno.land/r/nt/grc20reg/v0"
)

var Token *grc20.Token

func init(cur realm) {
	Token, _ = grc20.NewToken("Evil", "EVL", 4, 0, cur)
}

// TryRegister offers foo20's crossing functions beside this realm's token; the
// registry must refuse them.
func TryRegister(cur realm) {
	grc20reg.Register(cross(cur), Token, "", foo20.Transfer, foo20.Approve, foo20.TransferFrom)
}
