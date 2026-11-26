// PKGPATH: gno.land/r/zrealm_chalb_ownd
package zrealm_chalb_ownd

import "gno.land/r/tests/vm/chalbvictim"

func main(cur realm) {
	println("before:", chalbvictim.Total())
	chalbvictim.AddBoxD(cross(cur), 5)
	chalbvictim.AddBoxD(cross(cur), 5)
	println("after:", chalbvictim.Total())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/chalbvictim: an object written by read-only code would be stored (write it inside a crossing call of gno.land/r/tests/vm/chalbvictim)
