// PKGPATH: gno.land/r/zrealm_maker_handback
package zrealm_maker_handback

// The caller names the victim's own exported writer and hands it to the
// victim's gate: the caller made that handle, so it runs read-only there.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.SetOwner)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
