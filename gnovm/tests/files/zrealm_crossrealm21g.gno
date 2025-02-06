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
// cannot attach a reference to an unreal object from an external realm: heapitem((struct{(struct{("a" string)} gno.land/r/demo/tests/crossrealm/struct.A),(struct{("b" string)} gno.land/r/demo/tests/crossrealm/struct.B),(struct{( string)} gno.land/r/demo/tests/crossrealm/struct.D)} gno.land/r/demo/tests/crossrealm/struct.C))
