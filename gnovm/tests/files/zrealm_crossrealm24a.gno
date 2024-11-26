// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 *crossrealm.Bar = &crossrealm.Bar{A: 11} // not attached here

func init() {
	b0 = &crossrealm.Bar{A: 22}
}

func main() {
	print(".")
}

// Output:
// .
