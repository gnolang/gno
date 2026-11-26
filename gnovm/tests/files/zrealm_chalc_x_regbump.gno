// PKGPATH: gno.land/r/zrealm_chalc_x_regbump
package zrealm_chalc_x_regbump

import (
	"gno.land/r/tests/vm/chalcattack"
	"gno.land/r/tests/vm/chalcvictim"
)

func main(cur realm) {
	println("before:", chalcvictim.Total())
	chalcattack.RegBump(cross(cur))
	println("after:", chalcvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalcvictim's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
