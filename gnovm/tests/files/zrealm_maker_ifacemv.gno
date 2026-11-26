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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
