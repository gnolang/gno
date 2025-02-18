// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/array"
)

var root interface{}

func main() {
	root = crossrealm.GetArray2()
	println(".")
}

// Error:
// cannot attach a reference to an unattached object from an external realm
