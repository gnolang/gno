// PKGPATH: gno.land/r/launderrh_mv_returned_closure
package launderrh_mv_returned_closure

// /r/-victim returns a closure that captures gImm and writes through
// it. Attacker invokes the closure. Question: does borrow rule #1 fire on the
// closure body's /r/-victim source package and grant victim authority?
//
// 1. main → m.Realm = /r/-attacker
// 2. MakeWriterClosure("...") — borrow rule #1 → m.Realm = /r/-victim. Builds
//    closure (captures gImm), returns to attacker.
// 3. Attacker invokes closure() — borrow rule #1 fires (closure body
//    /r/-victim-declared) → m.Realm = /r/-victim during body.
// 4. Body: gImm.Field = value — write to /r/-victim-stamped under
//    victim authority → SUCCEEDS.
//
// Outcome: SUCCEEDS — this is by design (victim consented by
// returning a closure that writes). Sanity check on borrow rule #1 closure
// semantics.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	w := launderrvictim.MakeWriterClosure("pwnd-via-returned-closure")
	w()

	println("after: ", launderrvictim.ReadImm())
}

// Output:
// before: rdata-original
// after:  pwnd-via-returned-closure
