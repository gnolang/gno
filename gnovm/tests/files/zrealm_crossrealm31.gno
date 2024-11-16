// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice2()
	println(".")
}

// Error:
// should not happen while attempting to attach new real object from external realm
