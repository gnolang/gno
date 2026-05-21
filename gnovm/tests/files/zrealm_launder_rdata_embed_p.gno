// PKGPATH: gno.land/r/launderrh_embed
package launderrh_embed

// /r/-declared victim with EMBEDDED /p/-type — Apply-callback attack.
//
// WithEmbed (in /r/launderrvictim) embeds launderpkg.Immutable by
// value. Method promotion gives WithEmbed an Apply method (from
// /p/launderpkg). The attacker calls Apply with a /p/-declared
// function value as the callback.
//
// Trace:
//   1. main → m.Realm = /r/launderrh_embed
//   2. we.Apply(EvilWrite) — promoted method, dispatches to
//      (*Immutable).Apply on &we.Immutable.
//      recv's base is the WithEmbed struct, /r/launderrvictim-stamped
//      Rule 2 → m.Realm = /r/launderrvictim
//   3. Inside Apply: fn(i) calls EvilWrite (top-level /p/ fn).
//      No Rule 1 (not /r/), no Rule 2 (no receiver) → m.Realm stays
//      /r/launderrvictim
//   4. Inside EvilWrite: i.Field = "..." — write to
//      /r/launderrvictim-stamped data under /r/launderrvictim
//      authority → SUCCEEDS.
//
// This shows that embedding a /p/-type with Apply-style methods
// re-opens the H-class laundering, even when victim data is
// /r/-declared at the OUTER level.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	we.Apply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig
// after:  pwnd-via-apply
