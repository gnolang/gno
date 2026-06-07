// PKGPATH: gno.land/r/hivunrealborrow
package hivunrealborrow

// Regression test for the "var stopped bool in
// memberStorage.IterateByOffset" pattern (boards2/commondao).
//
// An unreal HIV is a transient heap-promotion wrapper for an escaping
// local — not a realm-owned slot. The readonly check skips the PkgID
// gate for unreal HIVs so writes to closure-captured locals are
// allowed across realm-borrow transitions.
//
// Trace for this test:
//
//   1. p_closurecap.Outer is a top-level /p/ function called from this
//      realm. No borrow on entry; m.Realm stays hivunrealborrow.
//      `var stopped bool` allocates an HIV with PkgID = this realm,
//      NewTime = 0 (unreal).
//   2. Outer immediately calls inn.loop(...). inn was obtained from
//      innerowner.GetInner(), so inn.PkgID = innerowner. loop is a
//      /p/ method on *Inner — PushFrameCall's borrow rule 2 shifts
//      m.Realm to innerowner for the duration of loop().
//   3. loop synchronously invokes the inline closure. The closure's
//      FuncLit is in /p/, so PushFrameCall does NOT shift m.Realm
//      back. The write `stopped = fn(i)` therefore evaluates its
//      LHS pointer under m.Realm = innerowner while HIV.PkgID =
//      hivunrealborrow.
//
// Without the unreal-HIV exception in IsReadonlyBy/isExternalRealm,
// the readonly check rejects the closure's write to its own captured
// slot. With it, the PkgID gate is skipped while NewTime == 0.

import (
	"gno.land/p/demo/tests/p_closurecap"
	"gno.land/r/tests/vm/innerowner"
)

func main(cur realm) {
	inn := innerowner.GetInner()

	// Stop at i==3 to confirm the captured `stopped` propagates.
	stopped := p_closurecap.Outer(inn, 10, func(i int) bool {
		return i == 3
	})
	println("stopped=", stopped)
}

// Output:
// stopped= true
