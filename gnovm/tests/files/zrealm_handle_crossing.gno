// PKGPATH: gno.land/r/zrealm_handle_crossing
package zrealm_handle_crossing

// A crossing function taken as a value still crosses when called with cur:
// the handed-back callback then runs read-only inside the victim's gate.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	run := cbvictim.RunS
	run(cross(cur), cbvictim.GateDecl(cross(cur)))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
