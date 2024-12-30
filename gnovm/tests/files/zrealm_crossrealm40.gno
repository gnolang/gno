// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm"

var f func() string

func main() {
	f = crossrealm.GetMethod() // XXX, this attach boundmethodvalue to current realm, shoud panic
	println(f)
}

// Error:
// cannot attach a value of a type defined by another realm
