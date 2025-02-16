// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/array"
)

var root interface{}

func main() {
	root = crossrealm.GetArray()
	println(".")
}

// Error:
// cannot attach a value of a type defined by another realm: struct{("a" string)}
