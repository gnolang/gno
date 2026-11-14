// PKGPATH: gno.land/r/zrealm_entry_iface_mkh
// A function value call runs view-only; the method value it returns has no
// stamp and acts for no realm when the victim calls it.
package zrealm_entry_iface_mkh

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	mk := ifaceattack.MakeH
	ifacevictim.ApplyOne(mk())
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
