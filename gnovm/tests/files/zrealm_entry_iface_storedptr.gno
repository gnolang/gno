// PKGPATH: gno.land/r/zrealm_entry_iface_storedptr
// A caller's library method stored as a pointer
// runs view-only: the realm does not import it.
package zrealm_entry_iface_storedptr

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.StoreRunner(cross(cur), &ifaceattack.EvilS{})
	pfuncvictim.UseStored(cross(cur))
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,1)>.Field
