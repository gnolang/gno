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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
