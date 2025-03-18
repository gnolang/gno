// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var f = func() bool {
	println("callback")
	return true
}

func main() {
	crossrealm.SetCallback(f)
	crossrealm.ExecuteCallback()
	println("ok")
}

// Output:
// callback
// ok

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/func"]
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm/func"]
// switchrealm["gno.land/r/crossrealm_test"]
