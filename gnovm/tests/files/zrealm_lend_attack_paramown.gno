// PKGPATH: gno.land/r/zrealm_lend_attack_paramown
package zrealm_lend_attack_paramown

import (
	"gno.land/r/tests/vm/lendpickattack"
	"gno.land/r/tests/vm/lendpickvictim"
)

func main(cur realm) {
	println("before:", lendpickvictim.Total())
	lendpickattack.ParamOwn(cross(cur))
	println("after:", lendpickvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendpickattack's code was entered without crossing and is read-only (call its crossing function): b<VPBlock(1,0)>.N
