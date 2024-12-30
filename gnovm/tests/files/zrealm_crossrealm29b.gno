// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

func main() {
	crossrealm.SetCallback(func() bool {
		println("callback")
		return true
	})
	crossrealm.ExecuteCallback()
}

// Error:
// cannot attach a value of a type defined by another realm
