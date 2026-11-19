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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
