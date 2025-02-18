// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

type Dt *crossrealm.Bar

var b Dt

func main() {
	b = &crossrealm.Bar{A: 22} // this should panic
	println(b)
}

// Error:
// cannot attach a value of a type defined by another realm
