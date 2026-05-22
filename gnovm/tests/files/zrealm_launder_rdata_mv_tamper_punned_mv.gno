// PKGPATH: gno.land/r/launderrh_mv_tamper_punned_mv
package launderrh_mv_tamper_punned_mv

// Type-punned bound method value laundering.
//
// 1. Attacker reads &we.Immutable (aliases /r/-victim memory).
// 2. Attacker punts pointer to *launderattack.Tamper via Convert
//    (same struct layout — pointer aliasing intact).
// 3. Attacker binds tamper.Set into a func(string) method value, then
//    stores it in /r/-attacker state.
// 4. Attacker invokes the stored mv from /r/-attacker context.
//    PushFrameCall for Tamper.Set sees recv stamped /r/-victim (the
//    underlying pointer alias is unchanged by the type pun) → Rule 2
//    → m.Realm = /r/-victim during Set body → write commits.
//
// Empirical outcome: BLOCKED at the Convert call site (the readonly
// taint propagates through type conversion). Storage doesn't get a
// chance to fire because the type pun itself is rejected. This shows
// the typepun defense fires regardless of subsequent storage shape.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

var stored func(string)

func main(cur realm) {
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	t := launderattack.Convert(&we.Immutable) // alias /r/-victim memory
	stored = t.Set
	stored("pwnd-via-stored-tamper-mv")

	println("after: ", launderrvictim.ReadEmbed())
}

// Error:
// illegal conversion of readonly or externally stored value
