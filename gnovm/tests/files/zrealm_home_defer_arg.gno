// PKGPATH: gno.land/r/zrealm_home_defer_arg
package zrealm_home_defer_arg

// A deferred call copies its arguments at the defer statement, under the
// deferring realm's hold; the deferred closure hands the value on.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	defer func(f func(string)) {
		cbvictim.RunS(cross(cur), f)
		println("owner:", cbvictim.Owner)
	}(cbvictim.FV)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
