// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var root interface{}

func main() {
	root = crossrealm.GetMap3()
	println("ok")
}

// Error:
// cannot attach objects by value from external realm: map{("a" string):(1 int)}
