// PKGPATH: gno.land/r/zrealm_chalc_x_paramown
package zrealm_chalc_x_paramown

import (
	"gno.land/r/tests/vm/chalcattack"
	"gno.land/r/tests/vm/chalcvictim"
)

func main(cur realm) {
	println("before:", chalcvictim.Total())
	chalcattack.ParamOwn(cross(cur))
	println("after:", chalcvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalcattack's code was entered without crossing and is read-only (call its crossing function): b<VPBlock(1,0)>.N
