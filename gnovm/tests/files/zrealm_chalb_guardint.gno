// PKGPATH: gno.land/r/zrealm_chalb_guardint
package zrealm_chalb_guardint

import (
	"gno.land/r/tests/vm/chalbattack"
	"gno.land/r/tests/vm/chalbvictim"
)

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbattack.GuardInt(cross(cur))
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalbvictim: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/chalbvictim)
