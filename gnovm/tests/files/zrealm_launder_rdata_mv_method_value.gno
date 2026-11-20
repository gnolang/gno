// PKGPATH: gno.land/r/launderrh_mv_method_value
package launderrh_mv_method_value

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Method-value laundering probe.
//
// Bind `Apply` method to a /r/-victim's WithEmbed receiver, save the
// method value into an /r/-attacker local variable, THEN invoke it
// later with the attacker's /p/-declared callback. Question:
//   Does the method value remember the recv stamp (/r/launderrvictim)?
//   When invoked from /r/-attacker context, does borrow rule #2 still fire?
//
// Trace expectation:
//   1. main → m.Realm = /r/launderrh_mv_method_value
//   2. mv := we.Apply — bound method value captures recv (we), recv
//      base is /r/launderrvictim-stamped.
//   3. mv(EvilWrite) — invoke later. PushFrameCall sees stamped recv,
//      borrow rule #2 → m.Realm = /r/launderrvictim
//   4. EvilWrite runs as top-level /p/ fn → no shift → write commits
//      under victim authority → SUCCEEDS.
//
// Outcome confirms method values preserve receiver borrowing.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	mv := we.Apply // bound method value
	mv(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadEmbed())
}

// Output:
// before: embed-orig

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
