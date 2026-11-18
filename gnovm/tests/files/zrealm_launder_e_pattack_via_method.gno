// PKGPATH: gno.land/r/laundere
package laundere

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Attack E: /p/attack.TamperViaMethod wraps the receiver-borrow
// loophole. The body of TamperViaMethod calls o.Set(s) — Set is a
// /p/-method on the victim-stamped *Object receiver, so borrow rule 2
// switches m.Realm to laundervictim for the duration of Set. Inside
// Set, the write to o.Field runs with victim authority and succeeds.
//
// Known-open loophole. The /p/attack package is a thin wrapper
// around the underlying capability — any /p/ package that imports
// a /p/ type and re-dispatches through that type's methods inherits
// the exploit. The borrow rule grants victim authority because the
// receiver lives in victim's state.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadPtr())
	p := laundervictim.GetPtr()
	launderattack.TamperViaMethod(p, "pwnd-via-pattack-method")
	println("after: ", laundervictim.ReadPtr())
}

// Output:
// before: original-ptr

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): o<VPBlock(1,0)>.Field
