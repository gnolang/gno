// PKGPATH: gno.land/r/curparity
package curparity

import (
	"chain/runtime"

	"gno.land/r/tests/vm/curchain_a"
)

// Parity invariant: for any execution context where both APIs are
// callable, `cur` (uverse interface) and `runtime.CurrentRealm()` /
// `runtime.PreviousRealm()` (chain/runtime struct) must agree on
// addr and pkgPath at the same chain position.
//
// Three contexts covered:
//
//	1. init()   — StageAdd entry. Row 1 of the parity table.
//	2. main()   — StageRun entry (MsgRun-shaped). Row 2.
//	3. Crossed-into helper (curchain_a.A) — cross-call shape; Row 3.

var (
	initOK, mainOK bool
)

func init(cur realm) {
	initOK = curMatchesRuntime(cur)
}

func main(cur realm) {
	mainOK = curMatchesRuntime(cur)

	println("init parity:", initOK)
	println("main parity:", mainOK)

	// Cross-call into curchain_a.A; A logs whether parity held there.
	curchain_a.A(cross)
}

// curMatchesRuntime checks that the captured `cur` value agrees on
// addr+pkgPath with what runtime.CurrentRealm() / runtime.PreviousRealm()
// would surface in the same execution context.
func curMatchesRuntime(cur realm) bool {
	rc := runtime.CurrentRealm()
	rp := runtime.PreviousRealm()
	if string(cur.Address()) != string(rc.Address()) {
		return false
	}
	if cur.PkgPath() != rc.PkgPath() {
		return false
	}
	prev := cur.Previous()
	if prev == nil {
		return false // we always expect a non-nil prev (origin realm)
	}
	if string(prev.Address()) != string(rp.Address()) {
		return false
	}
	if prev.PkgPath() != rp.PkgPath() {
		return false
	}
	return true
}

// Output:
// init parity: true
// main parity: true
// A cur.PkgPath: gno.land/r/tests/vm/curchain_a
// A prev.PkgPath: gno.land/r/curparity
// B cur.PkgPath: gno.land/r/tests/vm/curchain_b
// B Previous().PkgPath: gno.land/r/tests/vm/curchain_a
// B Previous().Previous().PkgPath: gno.land/r/curparity
// B Previous()^3 PkgPath empty (EOA origin): true
// B Previous()^4 panics past origin: true
// B Origin().PkgPath: gno.land/r/curparity
// B parity cur==Current: true
// B parity prev==Previous: true
