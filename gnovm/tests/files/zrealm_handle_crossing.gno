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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
