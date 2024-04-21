// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import "std"

var addr = std.Address("test")

func init() {
	addr = std.GetOrigCaller()
}

func main() {
	println(addr)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm

// Realm:
// switchrealm["gno.land/r/demo/tests_test"]
