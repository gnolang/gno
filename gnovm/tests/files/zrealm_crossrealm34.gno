// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var SS []*crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice5()
	println(".")
}

// Error:
// interface conversion: gnolang.PointerValue is not gnolang.Object: missing method DecRefCount
// *** CHECK THE ERR MESSAGES ABOVE, MAKE SURE IT'S WHAT YOU EXPECTED, DELETE THIS LINE AND RUN TEST AGAIN ***
