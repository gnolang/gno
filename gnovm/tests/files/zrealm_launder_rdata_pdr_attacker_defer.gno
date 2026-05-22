// PKGPATH: gno.land/r/launderrh_pdr_attacker_defer
package launderrh_pdr_attacker_defer

// Vector 6: deferred attacker write in main() runs after main()
// returns. The defer body is a /r/-attacker closure → Rule 1
// borrows m.Realm to attacker. m.Realm at defer time is whatever
// it was when main exited (attacker).
//
// Two variants:
//  A. Plain defer, no victim call interposed: m.Realm stays
//     attacker the whole time. Write to victim's gImm via
//     foreign-pointer should fail with readonly.
//  B. Defer captures a *Immutable pointer obtained BEFORE main's
//     body finishes. The captured pointer aliases victim memory.
//     Write attempt in defer body fires readonly.
//
// These are baseline checks: confirms there's no special "defer in
// main" privilege that bypasses the readonly check.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func try(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			println(label, "blocked: true")
		} else {
			println(label, "blocked: false")
		}
	}()
	fn()
}

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	// Variant A: try blocks each defer attempt synchronously
	try("A direct in attacker.main", func() {
		imm := launderrvictim.GetImm()
		imm.Field = "pwnd-A"
	})

	// Variant B: pointer captured outside, written inside a defer
	try("B defer write via captured ptr", func() {
		imm := launderrvictim.GetImm()
		// inner defer that writes
		defer func() {
			// This defer runs when the try-wrapped fn returns,
			// before try's outer defer/recover. m.Realm should
			// be attacker.
			imm.Field = "pwnd-B-defer"
		}()
		// Force the inner defer to run by simply returning.
	})

	// Variant C: synchronous victim call THEN defer write — does
	// the victim call leave m.Realm anywhere wrong?
	try("C victim-call then defer write", func() {
		_ = launderrvictim.ReadImm() // borrows + restores
		defer func() {
			imm := launderrvictim.GetImm()
			imm.Field = "pwnd-C-defer"
		}()
	})

	println("after: ", launderrvictim.ReadImm())
}

// Output:
// before: rdata-original
// A direct in attacker.main blocked: true
// B defer write via captured ptr blocked: true
// C victim-call then defer write blocked: true
// after:  rdata-original
