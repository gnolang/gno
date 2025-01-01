// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

// TODO: should panic at once
// var b0 crossrealm.Bar = crossrealm.Bar3
var root interface{}

func main() {
	root = crossrealm.Bar3
	println("ok")
}

// Error:
// cannot attach objects by value from external realm
