// PKGPATH: gno.land/r/zrealm_lend_box_keepstored
package zrealm_lend_box_keepstored

import (
	"gno.land/r/tests/vm/lendboxattack"
	"gno.land/r/tests/vm/lendboxvictim"
)

func main(cur realm) {
	println("before:", lendboxvictim.Total())
	lendboxattack.KeepStored(cross(cur))
	lendboxattack.KeepStored(cross(cur))
	println("after:", lendboxvictim.Total())
}

// Output:
// before: 0

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): b<VPBlock(1,0)>.N
