// PKGPATH: gno.land/r/demo/test_test
package test_test

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/tests"
)

func main() {
	var (
		eoa        = testutils.TestAddress("someone")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	std.TestSetOriginCaller(eoa)
	// Contrarily to z2_filetest.gno we EXPECT GetPreviousRealms != eoa (#1704)
	if addr := tests.GetPreviousRealm().Addr(); addr != eoa {
		println("want tests.GetPreviousRealm().Addr ==", eoa, "got", addr)
	}
	// When 2 or more realms in the frames, it is also different
	if addr := tests.GetRSubtestsPreviousRealm().Addr(); addr != rTestsAddr {
		println("want GetRSubtestsPreviousRealm().Addr ==", rTestsAddr, "got", addr)
	}
}

// Output:
// want tests.GetPreviousRealm().Addr == g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk got g1xufrdvnfk6zc9r0nqa23ld3tt2r5gkyvw76q63
