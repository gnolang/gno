// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

var f = func(s []crossrealm.XYZ) {
	SS = s
}

func main() {
	crossrealm.GetSlice4(f)
	println(".")
}

// Error:
// should not happen, struct{("0" string)} is not real
