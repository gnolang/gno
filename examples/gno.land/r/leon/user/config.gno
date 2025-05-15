package user

import (
	"gno.land/r/leon/factory"
)

var (
	Tok1 *factory.Token
	//Tok2 *factory.Token
)

// factory owns token

func init() {
	Tok1 = factory.NewToken("Leon")
	//Tok2 = cross(factory.NewTokenCrossing)("LeonCrossing")
}

// /r/leon/user    > g1le22u728yzatmw8zyxhwn4nh9tve5znyjw4r8e
// /r/leon/factory > g17e9vkl3wu0md8sjwdphjxjez99tv4zh7znep4a

func Render(_ string) string {
	out := "Tok1.Owner: " + Tok1.Owner.Owner().String()
	out += "\n\n"
	//out += "Tok2.Owner: " + Tok2.Owner.Owner().String()

	return out
}
