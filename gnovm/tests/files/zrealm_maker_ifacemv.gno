// PKGPATH: gno.land/r/zrealm_maker_ifacemv
package zrealm_maker_ifacemv

// An interface method value bound while read-only has no binder, so the
// victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GetIfaceMV())
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
