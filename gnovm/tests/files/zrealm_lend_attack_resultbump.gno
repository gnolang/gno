// PKGPATH: gno.land/r/zrealm_lend_attack_resultbump
package zrealm_lend_attack_resultbump

import (
	"gno.land/r/tests/vm/lendpickattack"
	"gno.land/r/tests/vm/lendpickvictim"
)

func main(cur realm) {
	println("before:", lendpickvictim.Total())
	lendpickattack.ResultBump(cross(cur))
	println("after:", lendpickvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendpickvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
