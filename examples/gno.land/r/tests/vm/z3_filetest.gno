// PKGPATH: gno.land/r/tests_z3
package tests_z3

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/tests/vm"
)

func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = chain.PackageAddress("gno.land/r/tests/vm")
	)
	testing.SetOriginCaller(eoa)
	// Contrarily to z2_filetest.gno we EXPECT GetPreviousRealms != eoa (#1704)
	if addr := vm.GetPreviousRealm(cross).Address(); addr != eoa {
		println("want vm.GetPreviousRealm().Address ==", eoa, "got", addr)
	}
	// When 2 or more realms in the frames, it is also different
	if addr := vm.GetRSubtestsPreviousRealm(cross).Address(); addr != rTestsAddr {
		println("want GetRSubtestsPreviousRealm().Address ==", rTestsAddr, "got", addr)
	}
	println("Done.")
}

// Output:
// Done.
