// PKGPATH: gno.land/r/demo/main

// SEND: 300ugnot

package main

import (
	"std"
	"testing"

	"gno.land/r/demo/disperse"
	tokens "gno.land/r/demo/grc20factory"
)

func main() {
	disperseAddr := std.DerivePkgAddr("gno.land/r/demo/disperse")
	mainaddr := std.DerivePkgAddr("gno.land/r/demo/main")
	beneficiary1 := std.Address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := std.Address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	testing.SetOriginCaller(mainaddr)

	banker := std.NewBanker(std.BankerTypeRealmSend)

	tokens.New("test", "TEST", 4, 0, 0)
	tokens.Mint("TEST", mainaddr, 200)

	mainbal := tokens.BalanceOf("TEST", mainaddr)
	println("main before:", mainbal)

	tokens.Approve("TEST", disperseAddr, 200)

	disperse.DisperseGRC20String("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "150TEST,50TEST")

	mainbal = tokens.BalanceOf("TEST", mainaddr)
	println("main after:", mainbal)
	ben1bal := tokens.BalanceOf("TEST", beneficiary1)
	println("beneficiary1:", ben1bal)
	ben2bal := tokens.BalanceOf("TEST", beneficiary2)
	println("beneficiary2:", ben2bal)
}

// Output:
// main before: 200
// main after: 0
// beneficiary1: 150
// beneficiary2: 50
