// PKGPATH: gno.land/r/zrealm_saved_variadic
// S3 attack: a fresh value packed into a variadic call in a later crossing
// call must not write.

package zrealm_saved_variadic

import "gno.land/r/tests/vm/cbsaved"

func main(cur realm) {
	f := cbsaved.MakeValue(cross(cur))
	cbsaved.CallVarg(cross(cur), f)
	println("owner:", cbsaved.Owner)
}

// Error:
// gno.land/r/tests/vm/cbsaved's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/cbsaved) package{})).Owner
