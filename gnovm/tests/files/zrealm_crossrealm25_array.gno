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
// references returned across realm boundaries must be real
