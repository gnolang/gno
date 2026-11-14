// PKGPATH: gno.land/r/zrealm_entry_iface_val
// A library method the caller picked, reached through the realm's interface,
// runs view-only: the realm's source never names it.
package zrealm_entry_iface_val

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.Use(ifaceattack.EvilV{})
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
