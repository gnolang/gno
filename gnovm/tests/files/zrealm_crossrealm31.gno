// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice2() // this is valid if elem are all attached in origin realm
	println(".")
}

// Error:
// should not happen while attempting to attach new real object from external realm
