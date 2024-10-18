// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 crossrealm.Bar

func init() {
	crossrealm.Bar2.A = 2 // this should be valid
}

func main() {
	print(".")
}
