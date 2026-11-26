// PKGPATH: gno.land/r/zrealm_chalb_keepintadder
package zrealm_chalb_keepintadder

import (
	"gno.land/r/tests/vm/chalbattack"
	"gno.land/r/tests/vm/chalbvictim"
)

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbattack.KeepIntAdder(cross(cur))
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalbvictim: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/chalbvictim)
