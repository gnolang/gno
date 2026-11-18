// PKGPATH: gno.land/r/zrealm_maker_handback
package zrealm_maker_handback

// The caller names the victim's own exported writer and hands it to the
// victim's gate: the caller made that function value, so it writes nothing.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.SetOwner)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/zrealm_maker_handback/zrealm_maker_handback.gno:10:28-45: cannot use non-crossing function gno.land/r/tests/vm/cbvictim.SetOwner of another realm as a value (call it, or wrap it in a closure)
