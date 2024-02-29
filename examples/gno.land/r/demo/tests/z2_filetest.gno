package main

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/tests"
)

func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	std.TestSetOrigCaller(eoa)
	// When a single realm in the frames, PrevRealm returns the user
	if addr := tests.GetPrevRealm().Addr(); addr != eoa {
		fmt.Printf("want tests.GetPrevRealm().Addr==%s, got %s\n", eoa, addr)
	}
	// When 2 or more realms in the frames, PrevRealm returns the second to last
	if addr := tests.GetRSubtestsPrevRealm().Addr(); addr != rTestsAddr {
		fmt.Printf("want GetRSubtestsPrevRealm().Addr==%s, got %s\n", rTestsAddr, addr)
	}
}

// Output:
