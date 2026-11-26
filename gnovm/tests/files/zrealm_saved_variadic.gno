// PKGPATH: gno.land/r/zrealm_saved_variadic
// R2a attack: a fresh value packed into a variadic call in a later crossing
// call must not write.

package zrealm_saved_variadic

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	f := cbsaved.MakeValue(cross(cur))
	cbsaved.CallVarg(cross(cur), f)
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
