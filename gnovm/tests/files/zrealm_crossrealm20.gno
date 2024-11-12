// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var b0 crossrealm.Bar // run declarations

func main() {
	print(".")
}

// Error:
// should not happen while attempting to attach new real object from external realm
