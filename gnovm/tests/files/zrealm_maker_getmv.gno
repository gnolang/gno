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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
