// PKGPATH: gno.land/r/zrealm_entry_iface_sub
// A caller's library method reached through an embedding interface
// runs view-only: the realm does not import it.
package zrealm_entry_iface_sub

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	ifacevictim.UseSub(ifaceattack.Evil(0))
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
