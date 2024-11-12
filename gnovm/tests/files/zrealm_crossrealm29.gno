// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

//var f1 = func() bool {
//	println("callback")
//	return true
//}

func main() {
	f := func() bool {
		println("callback")
		return true
	}
	crossrealm.SetCallback(f)
	crossrealm.ExecuteCallback()
}

// Error:
// should not happen while attempting to attach new real object from external realm
