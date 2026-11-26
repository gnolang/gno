// PKGPATH: gno.land/r/zrealm_chalc_x_fieldbump
package zrealm_chalc_x_fieldbump

import (
	"gno.land/r/tests/vm/chalcattack"
	"gno.land/r/tests/vm/chalcvictim"
)

func main(cur realm) {
	println("before:", chalcvictim.Total())
	chalcattack.FieldBump(cross(cur))
	println("after:", chalcvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalcvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
