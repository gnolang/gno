// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import (
	"gno.land/r/demo/tests"
	"std"
)

var addr = std.Address("test")
var addrInit = std.Address("addrInit")

func init() {
	addr = std.GetOrigCaller()
	addrInit = tests.InitOrigCaller()
}

func main() {
	println(addr)
	println(addrInit)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm

// Realm:
// switchrealm["gno.land/r/demo/tests_test"]
