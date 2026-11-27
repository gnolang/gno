// PKGPATH: gno.land/r/zrealm_entry_iface_storedval
// A struct value the realm stored is its own object, but the realm does not
// import the method's package, so the call still runs view-only.
package zrealm_entry_iface_storedval

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.StoreRunner(cross(cur), ifaceattack.EvilV{})
	pfuncvictim.UseStored(cross(cur))
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,1)>.Field
