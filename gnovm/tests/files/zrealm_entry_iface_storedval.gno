// PKGPATH: gno.land/r/zrealm_entry_iface_storedval
// Known open: a struct value the realm stores becomes its own object, so its
// library method later runs as the realm. This golden pins the gap.
package zrealm_entry_iface_storedval

import (
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.StoreRunner(pfuncbind.EvilV{})
	pfuncvictim.UseStored()
	println("victim:", pfuncvictim.Read())
}

// Output:
// victim: pwnd-via-iface
