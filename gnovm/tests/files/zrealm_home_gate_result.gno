// PKGPATH: gno.land/r/zrealm_home_gate_result
package zrealm_home_gate_result

// A value cbvictim's gate returns keeps its maker here, but this realm handing
// it back strips it.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	f := cbvictim.GateDecl(cross(cur))
	cbvictim.RunS(cross(cur), f)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
