// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

func main() {
	// even though we are running within a realm,
	// we aren't storing the result of crossrealm.Make1(),
	// so this should print fine.
	crossrealm.Make1().Touch().Print()
}

// Storage:
// gno.land/r/crossrealm: 0
// gno.land/r/tests/vm/crossrealm: 0

// Output:
// A: 2
// B: LocalStruct{123}
