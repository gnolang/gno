// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm"

var s *crossrealm.C

func cb(c *crossrealm.C) {
	s = c
}
func main() {
	crossrealm.GetStruct(cb)
	println(s)
}

// Error:
// ---should not happen, struct{("d" string)}: is not real
