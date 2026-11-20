// PKGPATH: gno.land/r/zrealm_t38_return_ptr
// T38: the realm's own writer sits behind a *func(); calling through the
// pointer is refused in the same crossing call and in a later one.

package zrealm_t38_return_ptr

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	p := &cbvictim.FV
	println("same:", revive(func() { cbvictim.RunP(cross(cur), p) }))
	cbvictim.RunP(cross(cur), p)
	println("owner:", cbvictim.Owner)
}

// Output:
// same: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
