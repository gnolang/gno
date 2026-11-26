// PKGPATH: gno.land/r/zrealm_lend_box_keepint
package zrealm_lend_box_keepint

import (
	"gno.land/r/tests/vm/lendboxattack"
	"gno.land/r/tests/vm/lendboxvictim"
)

func main(cur realm) {
	println("before:", lendboxvictim.Total())
	lendboxattack.KeepInt(cross(cur))
	println("after:", lendboxvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendboxvictim: an object written by read-only code would be stored (make the function that wrote it a crossing function of gno.land/r/tests/vm/lendboxvictim and call it with cross)
