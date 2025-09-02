package main

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/tests/tests"
)

// When a single realm in the frames, PreviousRealm returns the user
// When 2 or more realms in the frames, PreviousRealm returns the second to last
func main() {
	var (
		eoa = testutils.TestAddress("someone")
		_   = std.DerivePkgAddr("gno.land/r/tests/tests")
	)
	testing.SetOriginCaller(eoa)
	println("tests.GetPreviousRealm().Address(): ", tests.GetPreviousRealm(cross).Address())
	println("tests.GetRSubtestsPreviousRealm().Address(): ", tests.GetRSubtestsPreviousRealm(cross).Address())
}

// Output:
// tests.GetPreviousRealm().Address():  g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk
// tests.GetRSubtestsPreviousRealm().Address():  g1xstezhtvh0wsrj87mv3av55t4zj7zgla0x7pcq
