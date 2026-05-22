// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import (
	"chain/runtime/unsafe"

	"gno.land/r/tests/vm/crossrealm"
)

// Positive counterpart: external /r/-declared value held by caller;
// method dispatch borrows correctly to the declaring realm and the
// method body can read/inspect its own state without running afoul
// of the readonly taint.
//
// Make1 returns a *p_crossrealm.Container whose B field is a local
// /r/tests/vm/crossrealm.LocalStruct (PkgID=crossrealm). Calling
// LocalStruct.String() (an /r/-declared method) triggers Layer 1
// borrow to /r/tests/vm/crossrealm; inside the method, reading the
// receiver's own field succeeds.

func main(cur realm) {
	c := crossrealm.Make1()
	// c.B is *crossrealm.LocalStruct (an /r/-declared receiver).
	// Calling String() on it goes through Layer 1 borrow.
	println("B:", c.B.String())
	println("CurrentRealm after return:", unsafe.CurrentRealm().PkgPath())
}

// Output:
// B: LocalStruct{123}
// CurrentRealm after return: gno.land/r/caller
