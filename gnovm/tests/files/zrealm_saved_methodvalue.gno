// PKGPATH: gno.land/r/zrealm_saved_methodvalue
package zrealm_saved_methodvalue

// A method value stored in a package struct is read-only from the store on,
// also in the call that made it.

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	cbsaved.BindThenRun(cross(cur))
	println("box:", cbsaved.Box.Name)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name
