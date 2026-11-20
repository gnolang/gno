// PKGPATH: gno.land/r/zrealm_maker_methodvalue
package zrealm_maker_methodvalue

// The caller binds the victim's method on the victim's object: the binder is
// the caller, so the victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.Box.SetName)
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name
