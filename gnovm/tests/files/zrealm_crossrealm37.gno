// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm"

var m1 map[string]int

func main() {
	m1 = crossrealm.GetMap()
	println(m1)
}

// Error:
// should not happen while attempting to attach objects by value from external realm
