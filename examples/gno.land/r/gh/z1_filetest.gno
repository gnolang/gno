package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/gh"
)

var (
	adminAddr  = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	oracleAddr = std.Address("g1eunnckcl6r8ncwj0lrpxu9g5062xcvwxqlrf29")
	aliceAddr  = testutils.TestAddress("alice")
)

func main() {
	std.TestSetOrigCaller(oracleAddr)
	gh.OracleUpsertAccount("moul", "Manfred", "user")
	gh.OracleUpsertAccount("gnolang", "Gno Ecosystem", "org")
	gh.OracleUpsertRepo("gnolang/gno", false, false)
	moul := gh.AccountByID("moul")
	gnolang := gh.AccountByID("gnolang")
	gnolangGno := gh.RepoByID("gnolang/gno")
	println("gnoland:    ", gnolang.String())
	println("gnolangGno: ", gnolangGno.String())
	println("moul:       ", moul.String())
	println("Done.")
}

// Output:
// gnoland:     https://github.com/gnolang
// gnolangGno:  https://github.com/gnolang/gno
// moul:        https://github.com/moul
// Done.
