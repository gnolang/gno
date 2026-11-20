// PKGPATH: gno.land/r/launderrh_pdr_p_recover_pfn
package launderrh_pdr_p_recover_pfn

// Vector 8b: /p/ RecoverApply on a victim object runs the attacker's
// top-level /p/ EvilWrite through a value, so EvilWrite runs view-only and
// its write is refused. The view-only frame is no realm boundary, so
// RecoverApply's own recover catches the refusal and nothing was written.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadPtr())

	inner := launderrvictim.GetWithPtr().Inner
	rec := inner.RecoverApply(launderattack.EvilWrite)
	if rec == nil {
		println("rec: nil (no panic, write succeeded)")
	} else {
		println("rec: non-nil")
	}

	println("after: ", launderrvictim.ReadPtr())
}

// Output:
// before: ptr-orig
// rec: non-nil
// after:  ptr-orig
