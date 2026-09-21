// PKGPATH: gno.land/r/launderrh3
package launderrh3

// /r/-DECLARED victim data — closure callback laundering attempt.
//
// Victim exposes ApplyHook(h func(*Immutable)) which dispatches a
// caller-supplied callback on its /r/-declared data. The callback
// MUST be typed `func(*launderrvictim.Immutable)`, so /p/ packages
// (which can't import /r/) cannot provide it. Only an /r/ realm can.
//
// Attacker /r/launderrh3 provides a callback that writes the field.
// Trace:
//   1. main → m.Realm = /r/launderrh3
//   2. ApplyHook(fn) → borrow rule #1 → m.Realm = /r/launderrvictim
//   3. h(gImm) invokes fn (/r/launderrh3-declared closure)
//   4. Closure body → borrow rule #1 → m.Realm = /r/launderrh3
//   5. fn body: `i.Field = ...` writes /r/launderrvictim-stamped data
//      under /r/launderrh3 authority → readonly check fires

import (
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadImm())

	launderrvictim.ApplyHook(func(i *launderrvictim.Immutable) {
		i.Field = "pwnd-via-callback"
	})

	println("after: ", launderrvictim.ReadImm())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
