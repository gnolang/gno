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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
