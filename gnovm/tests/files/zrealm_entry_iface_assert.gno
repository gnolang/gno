// PKGPATH: gno.land/r/zrealm_entry_iface_assert
// A caller's library method reached by a type assertion
// runs view-only: the realm does not import it.
package zrealm_entry_iface_assert

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	ifacevictim.UseAssert(cross(cur), ifaceattack.Evil(0))
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
