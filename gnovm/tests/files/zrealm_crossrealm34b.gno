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
// cannot attach a value of a type defined by another realm
