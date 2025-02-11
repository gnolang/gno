// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

var f = func(s []crossrealm.XYZ) {
	SS = s
}

func main() {
	// XXX, even elem of the array is not real,
	// but the array is already attached to
	// external realm, so the elem has no
	// chance to attach to this realm.
	// so it isd valid.
	crossrealm.GetSlice4(f)
	println(".")
}

// Output:
// .
