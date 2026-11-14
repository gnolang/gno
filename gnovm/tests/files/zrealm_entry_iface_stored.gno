// PKGPATH: gno.land/r/zrealm_entry_iface_stored
package zrealm_entry_iface_stored

import (
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.StoreRunner(pfuncbind.Evil(0))
	pfuncvictim.UseStored()
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
