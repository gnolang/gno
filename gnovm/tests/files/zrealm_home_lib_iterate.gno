// PKGPATH: gno.land/r/zrealm_home_lib_iterate
package zrealm_home_lib_iterate

// cbvictim's gate hands its own value to a library that calls it by value:
// R2a refuses, the library call is a crossing call the value did not make.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunLibIterate(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbvictim) package{})).Owner
