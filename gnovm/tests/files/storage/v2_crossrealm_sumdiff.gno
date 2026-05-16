// PKGPATH: gno.land/r/bar
package bar

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

// interrealm v2 Phase 2: sumDiff routing — storage cost for objects accrue
// to their OWNING realm (PkgID), not the realm that triggered the
// save. Verifies touchedForeignRealms + FinalizeRealmTransaction's
// batch save / per-realm sumDiff drain.
//
// f originates in crossrealm_b (`Fooer` is a package-level var there).
// /r/bar calls crossrealm.SetFooer(f), which stores f's reference
// inside /r/tests/vm/crossrealm. f.SetS modifies f in place.
//
// Expected attribution:
// - /r/bar: 0 (caller realm; no state of its own changed)
// - /r/tests/vm/crossrealm: bytes for the reference + linkage  in its own state
// - /r/tests/vm/crossrealm_b: bytes for f's own struct fields (mutated SetS)

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(cross, f)
	f.SetS("ABCD")
	crossrealm.CallFooerFooCross(cross)
	print(".")
}

// Storage:
// gno.land/r/bar: 0
// gno.land/r/tests/vm/crossrealm: 214
// gno.land/r/tests/vm/crossrealm_b: 25

// Output:
// hello ABCD cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
