// PKGPATH: gno.land/r/zrealm_saved_methodval_fresh
// R2a attack: a method value of a fresh receiver made in one crossing call
// and invoked in a later one must not write.

package zrealm_saved_methodval_fresh

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	f := cbsaved.BindFresh(cross(cur), cbsaved.Box)
	cbsaved.CallBound(cross(cur), f)
	println("name:", cbsaved.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: t<VPBlock(1,0)>.Name
