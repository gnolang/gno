// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"chain/runtime"

	"gno.land/r/tests/vm/crossrealm"
)

func main(cur realm) {
	println(crossrealm.ExecCrossing(cross, func() string {
		rlm1 := runtime.CurrentRealm()
		rlm2 := runtime.PreviousRealm()
		return string(rlm1.Address()) + " " + rlm1.PkgPath() + "\n" + string(rlm2.Address()) + " " + rlm2.PkgPath()
	}))
}

// Output:
// g1kscyved3wru62zz4dpgncy7cs5wju8kcydzssk gno.land/r/tests/vm/crossrealm
// g1h2y7mn4d8w5ed08kqt8sdd7tp4j96eahyn6yan gno.land/r/crossrealm

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
