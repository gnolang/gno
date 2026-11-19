// PKGPATH: gno.land/r/launderrh_conv_iface_explicit
package launderrh_conv_iface_explicit

// CONVERSION VECTOR 2b: EXPLICIT cast to interface{}. Boxing holds no
// value, so the cast is accepted like the implicit assignment; the
// readonly mark travels with the pointer and the write is refused.

import (
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	w := launderrvictim.GetWithBarePtr()
	box := interface{}(w.Inner)
	box.(*launderpkg.Bare).Field = "pwnd-explicit-iface"
}

// Error:
// cannot directly modify readonly tainted object (call a crossing function of the owning realm): box<VPBlock(1,2)>.(*(typeval{gno.land/p/demo/tests/launderpkg/v0.Bare})).Field
