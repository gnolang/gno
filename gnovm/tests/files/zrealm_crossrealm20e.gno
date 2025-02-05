// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

var b1 *crossrealm.Bar

func main() {
	b1 = &crossrealm.Bar{A: 22} // this should panic
	println(b1)
}

// Error:
// cannot attach a reference to an unreal object from an external realm
