// PKGPATH: gno.land/r/tests/tests
package tests

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/tests/tests"
)

func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/tests/tests")
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

// Output:
// Done.
