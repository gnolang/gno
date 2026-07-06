// PKGPATH: gno.land/r/launderrh_pdr_p_recover_pfn
package launderrh_pdr_p_recover_pfn

// Vector 8b: /p/-method RecoverApply called with a top-level /p/
// callback (EvilWrite) — the known-open Apply attack, now inside
// a /p/-RecoverApply scope.
//
// Trace:
//  1. main → m.Realm = attacker
//  2. inner.RecoverApply(EvilWrite) → borrow rule #2 receiver-borrow on
//     *launderpkg.Immutable (victim-stamped) → m.Realm = victim
//  3. RecoverApply schedules defer { recover() } in /p/-frame
//  4. RecoverApply calls EvilWrite(i): EvilWrite is top-level /p/
//     fn → no borrow rule #1 (not /r/), no borrow rule #2 (no receiver) → m.Realm
//     STAYS victim
//  5. EvilWrite body: i.Field = ... → writes /r/-victim data under
//     m.Realm = victim → readonly check passes → WRITE SUCCEEDS
//  6. EvilWrite returns normally. No panic. Recover gets nil.
//
// Expected: known-open Apply attack succeeds; same outcome as
// pdr_p_defer_apply but synchronous (RecoverApply calls fn
// synchronously, only defer is the recover hook).

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	inner := launderrvictim.GetWithPtr().Inner
	rec := inner.RecoverApply(launderattack.EvilWrite)
	if rec == nil {
		println("rec: nil (no panic, write succeeded)")
	} else {
		println("rec: non-nil")
	}

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig
// rec: nil (no panic, write succeeded)
// after:  pwnd-via-apply
