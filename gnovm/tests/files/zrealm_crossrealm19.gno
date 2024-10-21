// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 = crossrealm.Bar{A: 1}

// un-attached package level var
// is this attached? or not? another function to attach?
//var b *crossrealm.Bar = crossrealm.SetBar(&b0)

func init() {
	//	crossrealm.CallFoo()
	crossrealm.SetBar(&b0)
}

func main() {
	crossrealm.CallBar()
	print(".")
}

// Output:
// hello gno.land/r/demo/tests/crossrealm
// hello gno.land/r/demo/tests/crossrealm
// .

// Error:

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
