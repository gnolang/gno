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
	"gno.land/p/nt/grc20/v0"
	"gno.land/r/nt/grc20reg/v0"
)

var (
	Token         *grc20.Token
	PrivateLedger *grc20.PrivateLedger
)

func init(cur realm) {
	// test20 only ever creates this one token, so id 0 can't collide.
	Token, PrivateLedger = grc20.NewToken("Test20", "TST", 4, 0, cur)
	grc20reg.Register(cross(cur), Token, "", Transfer, Approve, TransferFrom)
}

func Transfer(cur realm, to address, amount int64) {
	checkErr(PrivateLedger.CallerTeller().Transfer(0, cur, to, amount))
}

func Approve(cur realm, spender address, amount int64) {
	checkErr(PrivateLedger.CallerTeller().Approve(0, cur, spender, amount))
}

func TransferFrom(cur realm, from, to address, amount int64) {
	checkErr(PrivateLedger.CallerTeller().TransferFrom(0, cur, from, to, amount))
}

// Mint is intentionally unrestricted: test20 exists to fund fixtures.
func Mint(cur realm, to address, amount int64) {
	checkErr(PrivateLedger.Mint(to, amount))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
