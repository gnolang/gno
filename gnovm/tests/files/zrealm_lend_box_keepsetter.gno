// PKGPATH: gno.land/r/zrealm_lend_box_keepsetter
package zrealm_lend_box_keepsetter

import (
	"gno.land/r/tests/vm/lendboxattack"
	"gno.land/r/tests/vm/lendboxvictim"
)

func main(cur realm) {
	println("before:", lendboxvictim.Total())
	lendboxattack.KeepSetter(cross(cur))
	println("after:", lendboxvictim.Total())
}

// Output:
// before: 0

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,1)>.N
