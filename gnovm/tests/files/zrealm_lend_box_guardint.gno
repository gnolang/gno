// PKGPATH: gno.land/r/zrealm_lend_box_guardint
package zrealm_lend_box_guardint

import (
	"gno.land/r/tests/vm/lendboxattack"
	"gno.land/r/tests/vm/lendboxvictim"
)

func main(cur realm) {
	println("before:", lendboxvictim.Total())
	lendboxattack.GuardInt(cross(cur))
	println("after:", lendboxvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendboxvictim: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/lendboxvictim)
