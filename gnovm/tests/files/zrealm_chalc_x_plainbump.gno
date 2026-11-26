// PKGPATH: gno.land/r/zrealm_chalc_x_plainbump
package zrealm_chalc_x_plainbump

import (
	"gno.land/r/tests/vm/chalcattack"
	"gno.land/r/tests/vm/chalcvictim"
)

func main(cur realm) {
	println("before:", chalcvictim.Total())
	chalcattack.PlainBump(cross(cur))
	println("after:", chalcvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalcvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
