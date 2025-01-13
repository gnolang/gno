// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var m1 crossrealm.MyMap

func main() {
	m1 = crossrealm.GetMap()
	println(m1)
}

// Error:
// cannot attach objects by value from external realm
