// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	crossrealm "gno.land/r/tests/vm/crossrealm"
)

func main(cur realm) {
	// even though we are running within a realm,
	// we aren't storing the result of crossrealm.Make1(),
	// so this should print fine.
	crossrealm.Make1().Touch().Print()
}

// Output:
// A: 2
// B: LocalStruct{123}
