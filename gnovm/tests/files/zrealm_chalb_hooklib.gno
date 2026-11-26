// PKGPATH: gno.land/r/zrealm_chalb_hooklib
package zrealm_chalb_hooklib

import (
	"gno.land/r/tests/vm/chalbattack"
	"gno.land/r/tests/vm/chalbvictim"
)

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbattack.Hook(cross(cur))
	chalbvictim.RunHookLib(cross(cur))
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
