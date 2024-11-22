// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var SS [2]crossrealm.XYZ

var f = func(s [2]crossrealm.XYZ) {
	SS = s
}

func main() {
	crossrealm.GetSlice7(f)
	println(".")
}

// Error:
// should not happen while attempting to attach new real object from external realm
// *** CHECK THE ERR MESSAGES ABOVE, MAKE SURE IT'S WHAT YOU EXPECTED, DELETE THIS LINE AND RUN TEST AGAIN ***
