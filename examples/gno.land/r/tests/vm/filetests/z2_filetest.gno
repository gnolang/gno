package main

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	tests "gno.land/r/tests/vm"
)

// When a single realm in the frames, PreviousRealm returns the user
// When 2 or more realms in the frames, PreviousRealm returns the second to last
func main() {
	var (
		eoa = testutils.TestAddress("someone")
		_   = chain.PackageAddress("gno.land/r/tests/vm")
	)
	testing.SetOriginCaller(eoa)
	println("tests.GetPreviousRealm().Address(): ", tests.GetPreviousRealm(cross).Address())
	println("tests.GetRSubtestsPreviousRealm().Address(): ", tests.GetRSubtestsPreviousRealm(cross).Address())
}

// Output:
// tests.GetPreviousRealm().Address():  g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk
// tests.GetRSubtestsPreviousRealm().Address():  g1dhh6vhw9f5lmmpfz52rkf5dsk8lqzmad3fmpw7
