// PKGPATH: gno.land/r/zrealm_chalc_x_paramadder
package zrealm_chalc_x_paramadder

import (
	"gno.land/r/tests/vm/chalcattack"
	"gno.land/r/tests/vm/chalcvictim"
)

func main(cur realm) {
	println("before:", chalcvictim.Total())
	chalcattack.ParamAdder(cross(cur))
	println("after:", chalcvictim.Total())
}

// Output:
// before: 0

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,0)>.N
