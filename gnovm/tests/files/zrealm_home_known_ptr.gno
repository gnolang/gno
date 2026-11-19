// PKGPATH: gno.land/r/zrealm_home_known_ptr
package zrealm_home_known_ptr

// KNOWN (T20): cbvictim's gate reaches cbvictim's own value through a *func pointer
// this realm supplied; it cannot tell where the reference came from, so it writes.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunP(cross(cur), &cbvictim.FV)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: ptr
