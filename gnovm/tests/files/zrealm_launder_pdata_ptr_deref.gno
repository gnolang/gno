// PKGPATH: gno.land/r/launderpdata_p
package launderpdata_p

// /p/-source coverage: pointer-deref of a /p/-global from a /r/-realm.
//
// Companion to zrealm_launder_pdata_value_copy.gno but exercising the
// "deferred copy via pointer indirection" path: take address of a
// /p/-global, dereference in the /r/-frame to obtain a local copy,
// and try to mutate.
//
// Empirically the readonly check fires here too — same outcome as
// the value-read variant. The /p/-source bytes are not adoptable
// into /r/-authority via direct deref. The protection layer sits
// below the Copy gate.

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
