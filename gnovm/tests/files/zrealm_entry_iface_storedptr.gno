// PKGPATH: gno.land/r/zrealm_entry_iface_storedptr
package zrealm_entry_iface_storedptr

import (
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.StoreRunner(&pfuncbind.EvilS{})
	pfuncvictim.UseStored()
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
