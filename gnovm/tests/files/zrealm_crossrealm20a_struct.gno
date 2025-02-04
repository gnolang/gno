// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

var b0 crossrealm.Bar // run declarations

func main() {
	print(".")
}

// Error:
// cannot attach a value of a type defined by another realm
