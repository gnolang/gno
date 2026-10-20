// Package test20 implements a deliberately insecure ERC20 token for testing purposes.
// The Test20 token allows anyone to mint any amount of tokens to any address, making
// it unsuitable for production use. The primary goal of this package is to facilitate
// testing and experimentation without any security measures or restrictions.
//
//	WARNING: This token is highly insecure and should not be used in any
//	 production environment. It is intended solely for testing and
//	 educational purposes.
//
// It is built on gno.land/p/nt/grc20/v1 (uint256 amounts) and registered in
// gno.land/r/nt/grc20reg/v1.
package test20

import (
	"gno.land/p/nt/grc20/v1"
	"gno.land/r/nt/grc20reg/v1"
)

var (
	Token         *grc20.Token
	PrivateLedger *grc20.PrivateLedger
)

func init(cur realm) {
	// test20 only ever creates this one token, so id 0 can't collide.
	Token, PrivateLedger = grc20.NewToken("Test20", "TST", 4, 0, cur)
	grc20reg.Register(cross(cur), Token, "")
}
