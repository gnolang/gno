// PKGPATH: gno.land/r/launderpdata_p
package launderpdata_p

// /p/-source coverage: pointer-deref of a /p/-global from a /r/-realm.
//
// CLOSE_LAUNDER.md scenario 3 named this "the laundering trick" —
// passing a pointer across a realm boundary and letting the receiver
// deref-copy in its own frame to launder the stamp. The existing
// launder filetests didn't cover /p/-source variants.
//
// Empirically the readonly check fires here too — same outcome as the
// value-read variant. The /p/-source bytes are not adoptable into
// /r/-authority via direct deref, regardless of CLOSE_LAUNDER. The
// protection layer is below the Copy gate.

import (
	"gno.land/p/demo/tests/launderpkg"
)

func main(cur realm) {
	p := &launderpkg.PInitData
	stolen := *p
	stolen.Field = "pwnd-via-pdata-ptr-launder"
	println(stolen.Field)
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): stolen<VPBlock(1,2)>.Field
