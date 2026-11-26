// PKGPATH: gno.land/r/zrealm_lend_attack_paramlib
package zrealm_lend_attack_paramlib

import (
	"gno.land/r/tests/vm/lendpickattack"
	"gno.land/r/tests/vm/lendpickvictim"
)

func main(cur realm) {
	println("before:", lendpickvictim.Total())
	lendpickattack.ParamLib(cross(cur))
	println("after:", lendpickvictim.Total())
}

// Output:
// before: 0

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,0)>.N
