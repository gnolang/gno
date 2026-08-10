// PKGPATH: gno.land/r/launderpdata_p
package launderpdata_p

// /p/-source coverage: pointer-deref of a /p/-global from a /r/-realm.
//
// Post-#5747: dereferencing the pointer produces a fresh local
// struct copy owned by the executing realm. The local mutation
// succeeds and has no effect on the /p/-global. Direct writes
// *through* a /p/-stamped pointer (e.g. `p.Field = x` instead of
// `(*p).Field = x` on a deref-copy) remain blocked by the
// /p/-immutability gate.

import (
	"gno.land/p/demo/tests/launderpkg"
)

func main(cur realm) {
	p := &launderpkg.PInitData
	stolen := *p
	stolen.Field = "pwnd-via-pdata-ptr-launder"
	println(stolen.Field)
}

// Output:
// pwnd-via-pdata-ptr-launder
