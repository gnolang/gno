// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []*crossrealm.XYZ

var f = func(s []*crossrealm.XYZ) {
	SS = s
}

func main() {
	crossrealm.GetSlice6(f)
	println(".")
}

// Error:
// should not happen, struct{("6.1" string)} is not real
