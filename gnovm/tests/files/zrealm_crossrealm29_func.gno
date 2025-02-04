// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

func main() {
	f := func() bool {
		println("callback")
		return true
	}
	crossrealm.SetCallback(f)
	crossrealm.ExecuteCallback()
}

// Error:
// cannot attach a value of a type defined by another realm
