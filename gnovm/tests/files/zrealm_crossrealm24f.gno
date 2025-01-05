// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

var root interface{}

func main() {
	root = tests.F // already attached in p
	println("done")
}

// Error:
// cannot attach a value of a type defined by another realm
