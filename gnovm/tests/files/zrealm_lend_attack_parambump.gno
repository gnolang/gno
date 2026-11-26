// PKGPATH: gno.land/r/zrealm_lend_attack_parambump
package zrealm_lend_attack_parambump

import (
	"gno.land/r/tests/vm/lendpickattack"
	"gno.land/r/tests/vm/lendpickvictim"
)

func main(cur realm) {
	println("before:", lendpickvictim.Total())
	lendpickattack.ParamBump(cross(cur))
	println("after:", lendpickvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendpickvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
