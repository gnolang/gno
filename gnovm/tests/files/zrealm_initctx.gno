// PKGPATH: gno.land/r/demo/testinit
package testinit

import (
	"std"

	"gno.land/r/tests/tests"
)

var addr = std.Address("test")
var addrInit = std.Address("addrInit")

func init() {
	addr = std.OriginCaller()
	addrInit = tests.InitOriginCaller(cross)
}

func main(cur realm) {
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
// finalizerealm["gno.land/r/demo/testinit"]
