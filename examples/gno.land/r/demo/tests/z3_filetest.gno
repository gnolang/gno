// PKGPATH: gno.land/r/demo/test_test
package test_test

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
	if addr := cross(tests.GetPreviousRealm)().Address(); addr != eoa {
		println("want tests.GetPreviousRealm().Address ==", eoa, "got", addr)
	}
	// When 2 or more realms in the frames, it is also different
	if addr := cross(tests.GetRSubtestsPreviousRealm)().Address(); addr != rTestsAddr {
		println("want GetRSubtestsPreviousRealm().Address ==", rTestsAddr, "got", addr)
	}
	println("Done.")
}

// Output:
// Done.
