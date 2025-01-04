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
// cannot attach un-real object from external realm: struct{("0" string)}
