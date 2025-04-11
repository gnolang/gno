// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import (
	"chain"
	"chain/runtime"
	"gno.land/r/demo/tests"
)

var addr = chain.Address("test")
var addrInit = chain.Address("addrInit")

func init() {
	addr = runtime.OriginCaller()
	addrInit = tests.InitOriginCaller()
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
