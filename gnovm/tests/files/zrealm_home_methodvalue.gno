// PKGPATH: gno.land/r/zrealm_home_methodvalue
package zrealm_home_methodvalue

// cbvictim.MV is a method value cbvictim bound at init; handed back to its
// gate it has no binder, so it runs read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.MV)
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
