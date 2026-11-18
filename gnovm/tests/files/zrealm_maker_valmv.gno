// PKGPATH: gno.land/r/zrealm_maker_valmv
package zrealm_maker_valmv

// A value-receiver method value bound while read-only has no binder, so the
// victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GetValMV())
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.P.Name
