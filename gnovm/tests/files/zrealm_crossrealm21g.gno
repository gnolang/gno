// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/struct"

var s *crossrealm.C

func cb(c *crossrealm.C) {
	s = c
}
func main() {
	crossrealm.GetStruct2(cb)
	println(s)
}

// Error:
// cannot attach a value of a type defined by another realm
