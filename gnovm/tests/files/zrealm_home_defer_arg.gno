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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
