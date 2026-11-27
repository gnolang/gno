// PKGPATH: gno.land/r/launderrh_pdr_p_panic
package launderrh_pdr_p_panic

// Vector 11: /p/ PanicAfterApply runs the attacker's top-level /p/
// EvilWrite through a value, so EvilWrite runs view-only and its write is
// refused before the post-apply panic. The view-only frame is no realm
// boundary, so main's recover catches the refusal; nothing was written.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	defer func() {
		if r := recover(); r != nil {
			println("recovered:", r)
		} else {
			println("no recover")
		}
		println("after recover:", launderrvictim.ReadPtr())
	}()

	inner := launderrvictim.GetWithPtr().Inner
	inner.PanicAfterApply(launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig
// recovered: cannot modify object owned by another realm from library code (call a crossing function of the owning realm): i<VPBlock(1,0)>.Field
// after recover: ptr-orig
