// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 *crossrealm.Bar

func init() {
	b0 = &crossrealm.Bar{A: 1}
}

func main() {
	print(".")
}

// XXX, should work to attach by reference to
// type defined in another realm???

// Output:
// .
