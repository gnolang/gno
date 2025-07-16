// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
)

func main(cur realm) {
	println(crossrealm.ExecCrossing(cross, func() string {
		rlm1 := std.CurrentRealm()
		rlm2 := std.PreviousRealm()
		return string(rlm1.Address()) + " " + rlm1.PkgPath() + "\n" + string(rlm2.Address()) + " " + rlm2.PkgPath()
	}))
}

// Output:
// g1m6pzw9ex0xk3hamzch0wtad4e696gjyju0z2tz gno.land/r/demo/tests/crossrealm
// g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan gno.land/r/crossrealm

// Realm:
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
