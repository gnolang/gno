// PKGPATH: gno.land/r/demo/testinit
package testinit

import (
	"chain"
	"chain/runtime"

	"gno.land/r/demo/tests"
)

var addr = chain.Address("test")
var addrInit = chain.Address("addrInit")

func init() {
	addr = runtime.OriginCaller()
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
