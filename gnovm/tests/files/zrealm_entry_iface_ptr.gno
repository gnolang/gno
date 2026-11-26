// PKGPATH: gno.land/r/zrealm_entry_iface_ptr
// A caller's library method on a fresh pointer
// runs view-only: the realm does not import it.
package zrealm_entry_iface_ptr

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.Use(cross(cur), &ifaceattack.EvilS{})
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,1)>.Field
