// PKGPATH: gno.land/r/zrealm_handle_gate_own
package zrealm_handle_gate_own

// The victim's own declared function taken as a value inside its gate has
// the victim as maker: its gate runs it with write power.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GateDecl(cross(cur)))
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: picked
