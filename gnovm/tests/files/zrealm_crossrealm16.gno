// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 crossrealm.Bar

func init() {
	b0 = crossrealm.Bar{A: 1} // attach to this realm
}

func main() {
	crossrealm.SetBar(&b0) // set reference to external realm
	crossrealm.CallBar()
	print(".")
}

// Output:
// 1
// .

// Error:
