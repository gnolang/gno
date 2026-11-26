// PKGPATH: gno.land/r/zrealm_chalb_keepsetter
package zrealm_chalb_keepsetter

import (
	"gno.land/r/tests/vm/chalbattack"
	"gno.land/r/tests/vm/chalbvictim"
)

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbattack.KeepSetter(cross(cur))
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,1)>.N
