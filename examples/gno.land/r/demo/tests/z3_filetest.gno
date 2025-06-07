// PKGPATH: gno.land/r/demo/test_test
package tests

import (
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
	testing.SetOriginCaller(eoa)
	// Contrarily to z2_filetest.gno we EXPECT GetPreviousRealms != eoa (#1704)
	if addr := tests.GetPreviousRealm(cross).Address(); addr != eoa {
		println("want tests.GetPreviousRealm().Address ==", eoa, "got", addr)
	}
	// When 2 or more realms in the frames, it is also different
	if addr := tests.GetRSubtestsPreviousRealm(cross).Address(); addr != rTestsAddr {
		println("want GetRSubtestsPreviousRealm().Address ==", rTestsAddr, "got", addr)
	}
	println("Done.")
}

// Error:
// expected package name [test_test] but got [tests]

// TypeCheckError:
// gno.land/r/demo/test_test/.gnobuiltins.gno:1:1: package test_test; expected package tests; gno.land/r/demo/test_test/z3.gno:19:36: undefined: cross; gno.land/r/demo/test_test/z3.gno:23:45: undefined: cross
