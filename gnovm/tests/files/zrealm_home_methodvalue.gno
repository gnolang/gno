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
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
