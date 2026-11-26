// PKGPATH: gno.land/r/zrealm_lend_box_guardown
package zrealm_lend_box_guardown

import (
	"gno.land/r/tests/vm/lendboxattack"
	"gno.land/r/tests/vm/lendboxvictim"
)

func main(cur realm) {
	println("before:", lendboxvictim.Total())
	lendboxattack.GuardOwn(cross(cur))
	println("after:", lendboxvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendboxattack's code was entered without crossing and is read-only (call its crossing function): b<VPBlock(1,0)>.N
