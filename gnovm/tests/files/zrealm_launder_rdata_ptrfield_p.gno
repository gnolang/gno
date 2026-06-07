// PKGPATH: gno.land/r/launderrh_pf
package launderrh_pf

// /r/-declared victim with /p/-POINTER FIELD — Apply-callback attack.
//
// WithPtr has `Inner *launderpkg.Immutable`. Attacker reads the
// pointer field, then calls Apply with a /p/-declared callback.
//
// Same trace as embed variant: Apply borrows m.Realm to victim,
// /p/ callback inherits that authority, write succeeds.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	wp := launderrvictim.GetWithPtr()
	wp.Inner.Apply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig
// after:  pwnd-via-apply
