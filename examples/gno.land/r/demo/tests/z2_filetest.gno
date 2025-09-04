package main

import (
	"chain"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/tests"
)

// When a single realm in the frames, PreviousRealm returns the user
// When 2 or more realms in the frames, PreviousRealm returns the second to last
func main() {
	var (
		eoa = testutils.TestAddress("someone")
		_   = chain.DerivePkgAddr("gno.land/r/demo/tests")
	)
	testing.SetOriginCaller(eoa)
	println("tests.GetPreviousRealm().Address(): ", tests.GetPreviousRealm(cross).Address())
	println("tests.GetRSubtestsPreviousRealm().Address(): ", tests.GetRSubtestsPreviousRealm(cross).Address())
}

// Output:
// tests.GetPreviousRealm().Address():  g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk
// tests.GetRSubtestsPreviousRealm().Address():  g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
