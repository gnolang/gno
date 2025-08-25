// Package test20 implements a deliberately insecure ERC20 token for testing purposes.
// The Test20 token allows anyone to mint any amount of tokens to any address, making
// it unsuitable for production use. The primary goal of this package is to facilitate
// testing and experimentation without any security measures or restrictions.
//
//	WARNING: This token is highly insecure and should not be used in any
//	 production environment. It is intended solely for testing and
//	 educational purposes.
package test20

import (
	"gno.land/p/demo/tokens/grc20"
	"gno.land/r/demo/defi/grc20reg"
)

var Token, PrivateLedger = grc20.NewToken("Test20", "TST", 4)

func init() {
	grc20reg.Register(cross, Token, "")
}
