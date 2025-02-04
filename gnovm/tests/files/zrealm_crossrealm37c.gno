// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var root interface{}

func main() {
	root = crossrealm.GetMap4() // XXX, root is real, so check should happen in DidUpdate, namely, association panic
	println("ok")
}

// Error:
// cannot attach object by value from external realm: struct{}
