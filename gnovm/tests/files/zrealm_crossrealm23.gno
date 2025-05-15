// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
)

func main() {
	crossing()

	println(cross(crossrealm.ExecSR)(func() string {
		rlm1 := std.CurrentRealm()
		rlm2 := std.PreviousRealm()
		return string(rlm1.Address()) + " " + rlm1.PkgPath() + "\n" + string(rlm2.Address()) + " " + rlm2.PkgPath()
	}))
}

// Output:
// g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz gno.land/r/demo/tests/crossrealm
// g1vla5mffzum6060t99u4xhm8mnhgxr0sz4k574p gno.land/r/crossrealm_test

// Realm:
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm_test"]
