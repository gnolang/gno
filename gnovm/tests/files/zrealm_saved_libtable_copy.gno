// PKGPATH: gno.land/r/zrealm_saved_libtable_copy
package zrealm_saved_libtable_copy

// Library table called through a local copy: today it is already view-only.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.LibCopy(cross(cur))
	println("n:", cbsaved.LB.N)
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,0)>.N
