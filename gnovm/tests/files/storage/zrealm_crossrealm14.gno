// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

func main() {
	// even though we are running within a realm,
	// we aren't storing the result of crossrealm.Make1(),
	// so this should print fine.
	crossrealm.Make1().Touch().Print()
}

// Output:
// A: 2
// B: LocalStruct{123}

// Storage:
// gno.land/r/crossrealm_test: 0
// gno.land/r/demo/tests/crossrealm: 0

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
