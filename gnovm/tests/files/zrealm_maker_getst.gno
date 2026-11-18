// PKGPATH: gno.land/r/zrealm_maker_getst
package zrealm_maker_getst

// A method value bound while read-only, carried in a struct, has no binder
// either: the victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunSt(cross(cur), cbvictim.GetSt())
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
