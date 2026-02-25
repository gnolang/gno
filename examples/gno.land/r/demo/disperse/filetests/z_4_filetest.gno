// PKGPATH: gno.land/r/demo/main

package main

import (
	"chain"
	"testing"

	tokens "gno.land/r/demo/defi/grc20factory"
	"gno.land/r/demo/disperse"
)

func main() {
	disperseAddr := chain.PackageAddress("gno.land/r/demo/disperse")
	mainAddr := chain.PackageAddress("gno.land/r/demo/main")
	beneficiary1 := address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	testing.SetOriginCaller(mainAddr)

	tokens.New(cross, "test1", "TEST1", 4, 0, 0)
	tokens.Mint(cross, "TEST1", mainAddr, 200)
	println("main balance before (TEST1):", tokens.BalanceOf("TEST1", mainAddr))

	tokens.New(cross, "test2", "TEST2", 4, 0, 0)
	tokens.Mint(cross, "TEST2", mainAddr, 200)
	println("main balance before (TEST2):", tokens.BalanceOf("TEST2", mainAddr))

	tokens.Approve(cross, "TEST1", disperseAddr, 200)
	println("disperse allowance before (TEST1):", tokens.Allowance("TEST1", mainAddr, disperseAddr))

	tokens.Approve(cross, "TEST2", disperseAddr, 200)
	println("disperse allowance before (TEST2):", tokens.Allowance("TEST2", mainAddr, disperseAddr))

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	disperse.DisperseGRC20String(cross, addressesStr, "200TEST1,200TEST2")

	println("main balance after (TEST1):", tokens.BalanceOf("TEST1", mainAddr))
	println("main balance after (TEST2):", tokens.BalanceOf("TEST2", mainAddr))
	println("disperse allowance after (TEST1):", tokens.Allowance("TEST1", mainAddr, disperseAddr))
	println("disperse allowance after (TEST2):", tokens.Allowance("TEST2", mainAddr, disperseAddr))
	println("beneficiary1 balance (TEST1):", tokens.BalanceOf("TEST1", beneficiary1))
	println("beneficiary1 balance (TEST2):", tokens.BalanceOf("TEST2", beneficiary1))
	println("beneficiary2 balance (TEST1):", tokens.BalanceOf("TEST1", beneficiary2))
	println("beneficiary2 balance (TEST2):", tokens.BalanceOf("TEST2", beneficiary2))
}

// Output:
// main balance before (TEST1): 200
// main balance before (TEST2): 200
// disperse allowance before (TEST1): 200
// disperse allowance before (TEST2): 200
// main balance after (TEST1): 0
// main balance after (TEST2): 0
// disperse allowance after (TEST1): 0
// disperse allowance after (TEST2): 0
// beneficiary1 balance (TEST1): 200
// beneficiary1 balance (TEST2): 0
// beneficiary2 balance (TEST1): 0
// beneficiary2 balance (TEST2): 200
