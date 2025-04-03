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
	addrInit = withswitch(tests.InitOriginCaller)()
}

func main() {
	switchrealm()

	// Prints blanks because init's origin caller is blank.
	// XXX consider panic instead
	println(addr)
	println(addrInit)
}

// Output:
//
//

// Realm:
// switchrealm["gno.land/r/demo/tests_test"]
