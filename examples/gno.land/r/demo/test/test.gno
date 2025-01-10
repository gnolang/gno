package test

import (
	"gno.land/p/demo/grc/grc20"
	"gno.land/r/matijamarjanovic/tokenhub"
)

func Test() {
	token, _ := grc20.NewToken("Test123", "TEST123", 18)
	tokenhub.RegisterToken(token.Getter(), "test123")
}
