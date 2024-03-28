package main

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/tests"
)

// When a single realm in the frames, PrevRealm returns the user
// When 2 or more realms in the frames, PrevRealm returns the second to last
func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	std.TestSetOrigCaller(eoa)
	println("tests.GetPrevRealm().Addr(): ", tests.GetPrevRealm().Addr())
	println("tests.GetRSubtestsPrevRealm().Addr(): ", tests.GetRSubtestsPrevRealm().Addr())
}

// Output:
// tests.GetPrevRealm().Addr():  g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk
// tests.GetRSubtestsPrevRealm().Addr():  g1gz4ycmx0s6ln2wdrsh4e00l9fsel2wskqa3snq
