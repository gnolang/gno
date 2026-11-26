// PKGPATH: gno.land/r/zrealm_chalb_lit
package zrealm_chalb_lit

import (
	"gno.land/r/tests/vm/chalbattack"
	"gno.land/r/tests/vm/chalbvictim"
)

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbattack.RunLit(cross(cur))
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
