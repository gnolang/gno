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
	"gno.land/p/demo/grc/grc20"
	"gno.land/r/demo/grc20reg"
)

var Token, PrivateLedger = grc20.NewToken("Test20", "TST", 4)

func init() {
	cross(grc20reg.Register)(Token, "")
}
