// PKGPATH: gno.land/r/zrealm_handle_gate_own
package zrealm_handle_gate_own

// The victim's own declared function taken as a value inside its gate has
// the victim as maker; returned to another realm and handed back, it has none.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GateDecl(cross(cur)))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
