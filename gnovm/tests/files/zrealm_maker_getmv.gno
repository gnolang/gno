// PKGPATH: gno.land/r/zrealm_maker_getmv
package zrealm_maker_getmv

// A method value the victim's helper binds while read-only has no binder, so
// the victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.GetMV())
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
