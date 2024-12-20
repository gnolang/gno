package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/tests"
)

// When a single realm in the frames, PreviousRealm returns the user
// When 2 or more realms in the frames, PreviousRealm returns the second to last
func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	std.TestSetOriginCaller(eoa)
	println("tests.GetPreviousRealm().Addr(): ", tests.GetPreviousRealm().Addr())
	println("tests.GetRSubtestsPreviousRealm().Addr(): ", tests.GetRSubtestsPreviousRealm().Addr())
}

// Output:
// tests.GetPreviousRealm().Addr():  g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk
// tests.GetRSubtestsPreviousRealm().Addr():  g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
