// PKGPATH: gno.land/r/demo/main

package main

import (
	"std"
	"testing"

	"gno.land/r/demo/disperse"
	tokens "gno.land/r/demo/grc20factory"
)

func main() {
	disperseAddr := std.DerivePkgAddr("gno.land/r/demo/disperse")
	mainAddr := std.DerivePkgAddr("gno.land/r/demo/main")
	beneficiary1 := std.Address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := std.Address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	testing.SetOriginCaller(mainAddr)

	tokens.New(cross, "test", "TEST", 4, 0, 0)
	tokens.Mint(cross, "TEST", mainAddr, 200)
	println("main balance before:", tokens.BalanceOf("TEST", mainAddr))

	tokens.Approve(cross, "TEST", disperseAddr, 200)
	println("disperse allowance before:", tokens.Allowance("TEST", mainAddr, disperseAddr))

	addressesStr := beneficiary1.String() + "," + beneficiary2.String()
	disperse.DisperseGRC20String(cross, addressesStr, "150TEST,50TEST")

	println("main balance after:", tokens.BalanceOf("TEST", mainAddr))
	println("disperse allowance after:", tokens.Allowance("TEST", mainAddr, disperseAddr))
	println("beneficiary1 balance:", tokens.BalanceOf("TEST", beneficiary1))
	println("beneficiary2 balance:", tokens.BalanceOf("TEST", beneficiary2))
}

// Output:
// main balance before: 200
// disperse allowance before: 200
// main balance after: 0
// disperse allowance after: 0
// beneficiary1 balance: 150
// beneficiary2 balance: 50
