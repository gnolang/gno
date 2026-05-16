// PKGPATH: gno.land/r/demo/testinit
package testinit

import (
	"chain/runtime"

	tests "gno.land/r/tests/vm"
)

var (
	addr     = address("test")
	addrInit = address("addrInit")
)

func init(cur realm) {
	addr = runtime.OriginCaller()
	addrInit = tests.InitOriginCaller(cross2(cur))
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
