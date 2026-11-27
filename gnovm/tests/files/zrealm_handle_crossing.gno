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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
