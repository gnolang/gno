// PKGPATH: gno.land/r/zrealm_entry_iface_gen
// A method value the caller's library built while view-only has no stamp, so
// it acts for no realm when the victim calls it.
package zrealm_entry_iface_gen

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	ifacevictim.UseGen(cross(cur), ifaceattack.Gen{})
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
