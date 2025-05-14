// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import (
	"std"

	"gno.land/r/demo/tests"
)

var addr = std.Address("test")
var addrInit = std.Address("addrInit")

func init() {
	addr = std.OriginCaller()
	addrInit = cross(tests.InitOriginCaller)()
}

func main() {
	crossing()

	// Prints blanks because init's origin caller is blank.
	// XXX consider panic instead
	println(addr)
	println(addrInit)
	println(".") // trailing newline is needed.
}

// Output:
//
//
// .

// Realm:
// finalizerealm["gno.land/r/demo/tests_test"]
