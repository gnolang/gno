// PKGPATH: gno.land/r/demo/main

// GRC20 amounts are uint256: dispersing more than an int64 can hold works,
// and an amount that is not a decimal number is refused.
package main

import (
	"chain"
	"testing"

	tokens "gno.land/r/demo/defi/grc20factory"
	"gno.land/r/demo/disperse"
)

func main(cur realm) {
	disperseAddr := chain.PackageAddress("gno.land/r/demo/disperse")
	mainAddr := chain.PackageAddress("gno.land/r/demo/main")
	beneficiary1 := address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	testing.SetOriginCaller(mainAddr)

	big := uint256(1) << 100
	tokens.New(cross(cur), "big", "BIG", 18, 0, 0)
	tokens.Mint(cross(cur), "BIG", mainAddr, 2*big)
	tokens.Approve(cross(cur), "BIG", disperseAddr, 2*big)

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	// 2^100 = 1267650600228229401496703205376
	disperse.DisperseGRC20String(cross(cur), addressesStr,
		"1267650600228229401496703205376BIG,1267650600228229401496703205376BIG")

	println("main balance after:", tokens.BalanceOf("BIG", mainAddr))
	println("beneficiary1 balance:", tokens.BalanceOf("BIG", beneficiary1))
	println("beneficiary2 balance:", tokens.BalanceOf("BIG", beneficiary2))

	disperse.DisperseGRC20String(cross(cur), beneficiary1.String(), "-1BIG")
}

// Output:
// main balance after: 0
// beneficiary1 balance: 1267650600228229401496703205376
// beneficiary2 balance: 1267650600228229401496703205376

// Error:
// disperse: invalid amount
