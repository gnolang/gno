// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"std"

	"gno.land/r/tests/crossrealm"
)

func main(cur realm) {
	println(crossrealm.ExecCrossing(cross, func() string {
		rlm1 := std.CurrentRealm()
		rlm2 := std.PreviousRealm()
		return string(rlm1.Address()) + " " + rlm1.PkgPath() + "\n" + string(rlm2.Address()) + " " + rlm2.PkgPath()
	}))
}

// Output:
// g1wtde2f2gm29n4k3pdtzjqc4pqeyqj9upgwnsch gno.land/r/tests/crossrealm
// g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan gno.land/r/crossrealm

// Realm:
// finalizerealm["gno.land/r/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
