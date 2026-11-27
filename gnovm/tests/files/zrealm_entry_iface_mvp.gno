// PKGPATH: gno.land/r/zrealm_entry_iface_mvp
// A caller's library method taken as a method value (pointer receiver)
// runs view-only: the realm does not import it.
package zrealm_entry_iface_mvp

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	ifacevictim.UseMV(cross(cur), &ifaceattack.EvilS{})
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,1)>.Field
