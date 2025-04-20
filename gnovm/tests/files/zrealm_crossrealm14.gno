// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

func main() {
	crossing()

	// even though we are running within a realm,
	// we aren't storing the result of crossrealm.Make1(),
	// so this should print fine.
	crossrealm.Make1().Touch().Print()
}

// Output:
// A: 2
// B: LocalStruct{123}
