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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
