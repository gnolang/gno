package tokenhub

import (
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	tokenA *grc20.Token
	tokenB *grc20.Token
	tokenC *grc20.Token
	tokenD *grc20.Token
	tokenE *grc20.Token
)

func init() {
	tokenA = grc20.NewToken("Token A", "TKNA", 6)
	tokenB = grc20.NewToken("Token B", "TKNB", 6)
	tokenC = grc20.NewToken("Token C", "TKNC", 6)
	tokenD = grc20.NewToken("Token D", "TKND", 6)
	tokenE = grc20.NewToken("Token E", "TKNE", 6)

	Send5OfEach("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	RegisterToken(tokenA)
	RegisterToken(tokenB)
	RegisterToken(tokenC)
	RegisterToken(tokenD)
	RegisterToken(tokenE)
}

func Send5OfEach(to std.Address) error {
	if !to.IsValid() {
		return grc20.ErrInvalidAddress
	}

	if err := adminA.Mint(to, 5); err != nil {
		return err
	}
	if err := adminB.Mint(to, 5); err != nil {
		return err
	}
	if err := adminC.Mint(to, 5); err != nil {
		return err
	}
	if err := adminD.Mint(to, 5); err != nil {
		return err
	}
	if err := adminE.Mint(to, 5); err != nil {
		return err
	}

	return nil
}
