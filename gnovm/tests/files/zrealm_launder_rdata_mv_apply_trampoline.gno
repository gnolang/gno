// PKGPATH: gno.land/r/launderrh_mv_apply_trampoline
package launderrh_mv_apply_trampoline

// Victim returns a "packaged Apply" — a closure of type
// `func(func(*launderrvictim.Immutable))`. The body is /r/-victim-
// declared. Because the callback's parameter type names the
// /r/-victim's Immutable type, /p/ packages cannot supply the
// callback (can't import /r/). Only /r/-realms can — and any
// /r/-attacker callback body triggers its own Rule 1 shift,
// blocking the write.
//
// Trace:
//   1. main → m.Realm = /r/-attacker
//   2. MakeApplyTrampoline() — Rule 1 → /r/-victim, returns closure
//      cap = (gImm).
//   3. Attacker invokes cap(writeVic) — Rule 1 (closure body) →
//      m.Realm = /r/-victim during body.
//   4. Body: fn(gImm) calls writeVic. writeVic is /r/-attacker-
//      declared → Rule 1 → m.Realm = /r/-attacker.
//   5. writeVic body: i.Field = "..." — readonly check fires →
//      BLOCKED.
//
// Outcome: BLOCKED — using /r/-victim-typed callback parameter
// closes the /p/-callback laundering door at the trampoline layer,
// the same way ApplyHook (with `func(*Immutable)` argument type)
// closes it for the direct call form.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

// Helper /r/-attacker-declared callback (the closure must have the
// concrete victim-pointer type, which /p/ packages can't name).
func writeVic(i *launderrvictim.Immutable) {
	i.Field = "pwnd-via-trampoline-rh-cb"
}

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	apply := launderrvictim.MakeApplyTrampoline()
	apply(writeVic)

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
