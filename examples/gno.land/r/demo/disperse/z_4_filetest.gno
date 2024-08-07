// SEND: 300ugnot

package main

import (
	"std"

	"gno.land/r/demo/disperse"
	tokens "gno.land/r/demo/grc20factory"
)

func main() {
	disperseAddr := std.DerivePkgAddr("gno.land/r/demo/disperse")
	mainaddr := std.DerivePkgAddr("main")
	beneficiary1 := std.Address("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0")
	beneficiary2 := std.Address("g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c")

	std.TestSetOrigPkgAddr(disperseAddr)
	std.TestSetOrigCaller(mainaddr)

	banker := std.GetBanker(std.BankerTypeRealmSend)

	tokens.New("test1", "TEST1", 4, 0, 0)
	tokens.Mint("TEST1", mainaddr, 200)
	tokens.New("test2", "TEST2", 4, 0, 0)
	tokens.Mint("TEST2", mainaddr, 200)

	mainbal := tokens.BalanceOf("TEST1", mainaddr) + tokens.BalanceOf("TEST2", mainaddr)
	println("main before:", mainbal)

	tokens.Approve("TEST1", disperseAddr, 200)
	tokens.Approve("TEST2", disperseAddr, 200)

	disperse.DisperseGRC20String("g1dmt3sa5ucvecxuhf3j6ne5r0e3z4x7h6c03xc0,g1akeqsvhucjt8gf5yupyzjxsjd29wv8fayng37c", "200TEST1,200TEST2")

	mainbal = tokens.BalanceOf("TEST1", mainaddr) + tokens.BalanceOf("TEST2", mainaddr)
	println("main after:", mainbal)
	ben1bal := tokens.BalanceOf("TEST1", beneficiary1) + tokens.BalanceOf("TEST2", beneficiary1)
	println("beneficiary1:", ben1bal)
	ben2bal := tokens.BalanceOf("TEST1", beneficiary2) + tokens.BalanceOf("TEST2", beneficiary2)
	println("beneficiary2:", ben2bal)
}

// Output:
// main before: 400
// main after: 0
// beneficiary1: 200
// beneficiary2: 200
