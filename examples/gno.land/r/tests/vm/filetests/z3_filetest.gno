// PKGPATH: gno.land/r/tests
package tests

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	tests "gno.land/r/tests/vm"
)

func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = chain.PackageAddress("gno.land/r/tests/vm")
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
