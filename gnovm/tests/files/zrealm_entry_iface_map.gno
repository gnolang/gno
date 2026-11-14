// PKGPATH: gno.land/r/zrealm_entry_iface_map
// A caller's library method read from a map
// runs view-only: the realm does not import it.
package zrealm_entry_iface_map

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	ifacevictim.UseMap(map[string]ifacevictim.Runner{"a": ifaceattack.Evil(0)})
	println("victim:", ifacevictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): x<VPBlock(1,1)>.Field
