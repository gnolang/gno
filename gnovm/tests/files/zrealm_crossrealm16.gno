// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 crossrealm.Bar

func init() {
	b0 = crossrealm.Bar{A: 1} // XXX, should be invalid to attach external type
}

func main() {
	crossrealm.SetBar(&b0)
	crossrealm.CallBar()
	print(".")
}
