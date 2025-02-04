// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var root interface{}

func main() {
	root = crossrealm.GetMap4()
	println("ok")
}

// Error:
// cannot attach a value of a type defined by another realm
