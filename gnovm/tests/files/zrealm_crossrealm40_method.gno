// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/func"

var f func() string

func main() {
	f = crossrealm.GetMethod() // XXX, this attach boundmethodvalue to current realm, shoud panic
	println(f)
}

// Output:
// <*gno.land/r/demo/tests/crossrealm/func.MyStruct>.M(sv *gno.land/r/demo/tests/crossrealm/func.MyStruct)( string)
