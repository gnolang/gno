// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
)

var root interface{}

func main() {
	//f := func() crossrealm.Fooer {
	root = func() crossrealm.Fooer {
		return crossrealm.Fooer(nil)
	}
	println("ok")
}

// Error:
// cannot attach function contains object from external realm
