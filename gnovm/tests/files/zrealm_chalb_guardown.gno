// PKGPATH: gno.land/r/zrealm_chalb_guardown
package zrealm_chalb_guardown

import (
	"gno.land/r/tests/vm/chalbattack"
	"gno.land/r/tests/vm/chalbvictim"
)

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbattack.GuardOwn(cross(cur))
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalbattack's code was entered without crossing and is read-only (call its crossing function): b<VPBlock(1,0)>.N
