// PKGPATH: gno.land/r/zrealm_entry_iface_embedptr
// A caller's library method promoted through a pointer to an embedding struct
// runs view-only: the realm does not import it.
package zrealm_entry_iface_embedptr

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	ifacevictim.UseEmbedPtr(cross(cur), &ifaceattack.EvilS{})
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
