// PKGPATH: gno.land/r/demo/test_test
package test_test

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
	// Contrarily to z2_filetest.gno we EXPECT GetPrevRealms != eoa (#1704)
	if addr := tests.GetPrevRealm().Addr(); addr != eoa {
		println("want tests.GetPrevRealm().Addr ==", eoa, "got", addr)
	}
	// When 2 or more realms in the frames, it is also different
	if addr := tests.GetRSubtestsPrevRealm().Addr(); addr != rTestsAddr {
		println("want GetRSubtestsPrevRealm().Addr ==", rTestsAddr, "got", addr)
	}
}

// Output:
// want tests.GetPrevRealm().Addr == g1wdhk6et0dej47h6lta047h6lta047h6lrnerlk got g1xufrdvnfk6zc9r0nqa23ld3tt2r5gkyvw76q63
