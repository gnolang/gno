// PKGPATH: gno.land/r/zrealm_home_registry
package zrealm_home_registry

// Intended: cbvictim's own non-crossing callback kept by a third realm's
// registry comes back makerless, so it writes nothing; registries take crossing functions.

import "gno.land/r/tests/vm/cbvictim"

var reg func(string)

func put(cur realm, f func(string)) { reg = f }
func get(cur realm) func(string)    { return reg }

func main(cur realm) {
	cbvictim.RegVia(cross(cur), put, get)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
