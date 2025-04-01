// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
)

func main() {

	println(withswitch(crossrealm.SRWithSwitch)(func() string {
		rlm1 := std.CurrentRealm()
		rlm2 := std.PreviousRealm()
		return string(rlm1.Address()) + " " + rlm1.PkgPath() + "\n" + string(rlm2.Address()) + " " + rlm2.PkgPath()
	}))
}

// Output:
// g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz gno.land/r/demo/tests/crossrealm
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
