// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice2()
	println(".")
}

// Error:
// references returned across realm boundaries must be real
