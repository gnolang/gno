// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var f = func() bool {
	println("callback")
	return true
}

func main() {
	crossrealm.SetCallback(f) // func value
	crossrealm.ExecuteCallback()
}

// Error:
// cannot attach objects by value from external realm
